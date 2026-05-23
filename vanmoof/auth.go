// auth.go — port of bleware/auth.c session-key derivation and the
// supporting keyed-record API from bleware/secrets.c.
//
// The BLE firmware's `auth_derive_session_key` (@ 0x00018B1C) resolves
// the 32-byte session-key record for a client-supplied 32-bit key id:
//
//  1. Linear-scan slots [0, 0x7B] of the SPI-flash secrets sector for
//     a CRC-valid record whose payload+16 word equals client_key_id
//     (see SecretsFindByKey / secrets_find_by_key @ 0x00022BAA).
//  2. On a hit, return the matching record verbatim.
//  3. On a miss, *if* the device is unprovisioned — no CRC-valid
//     manufacturing key at slot 0x7E and zero CRC-valid records in
//     [0, 0x7B] — synthesise and return a default OWNER_PERMS
//     record. Otherwise fail.
//
// This file ports both paths so a host tool can replay the exact
// session-key resolution against an SPI-flash dump.
//
// Record layout used by the keyed-record API (32 bytes, from
// secrets.c — supersedes the looser layout description in auth.c):
//
//   +0x00  16 B  application payload  (e.g. "_____OWNER_PERMS")
//   +0x10   4 B  key            (uint32 LE, matched by find_by_key)
//   +0x14   4 B  application-defined (OWNER_PERMS stores 0xFFFFFFFF
//                here — auth.c calls it a permission mask)
//   +0x18   4 B  tag            (uint32 LE, "UKEY" / "M-ID" / ...)
//   +0x1C   4 B  CRC-32/LE of bytes 0..0x1B
//
// CRC variant: OEM `crc32_le` (reflected, seed 0xFFFFFFFF, no final
// XOR) — see CalcCRC32LE in crc.go.

package vanmoof

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Known client key ids (the uint32 at payload+16). The OEM Fixie/
// VanMoof app uses these slots; mapping is from the "Features" list
// in README.md and from observed dumps.
var clientKeyNames = map[uint32]string{
	0x00000000: "OWNER_PERMS (default)",
	0x00000001: "BikeComm",
	0x00000002: "Sharing",
	0x00000003: "Workshop",
}

// clientKeyName returns a human-readable label for a known client
// key id, or "unknown" for ids we have not catalogued yet.
func clientKeyName(id uint32) string {
	if name, ok := clientKeyNames[id]; ok {
		return name
	}
	return "unknown"
}

// Slot layout of the external-flash secrets sector. Slots [0, 0x7B]
// are the user-keyed range (`SECRETS_KEYED_SLOT_LIMIT == 0x7C` in
// secrets.c); slots [0x7C, 0x7F] are reserved for directory and
// manufacturing-related records.
const (
	// secretsKeysRegionFirst is the first slot index in the keys
	// region.
	secretsKeysRegionFirst = 0x00

	// secretsKeysRegionLast is the inclusive last slot index of the
	// user-keyed region (0x7B == 123). secrets.c defines the matching
	// exclusive limit as SECRETS_KEYED_SLOT_LIMIT == 0x7C.
	secretsKeysRegionLast = 0x7B

	// secretsMidRecordSlot is the M-ID directory slot (124). Holds a
	// CRC-record tagged "M-ID" — see secrets_ensure_mid_record.
	secretsMidRecordSlot = 0x7C

	// secretsMfgKeySlot is the slot the firmware probes with
	// `secrets_record_read(0x7E, NULL)` in auth_derive_session_key to
	// decide whether a manufacturing key has been written.
	secretsMfgKeySlot = 0x7E

	// secretsKeyOffset is the byte offset within a record's payload
	// where the 32-bit lookup key lives. Equals SECRETS_KEY_OFFSET
	// from secrets.c.
	secretsKeyOffset = 0x10

	// secretsTagOffset is the byte offset within a record's payload
	// where the 4-byte record-type tag lives. Equals
	// SECRETS_TAG_OFFSET from secrets.c.
	secretsTagOffset = 0x18

	// secretsTagUKey is the little-endian "UKEY" tag stamped at
	// payload+24 by secrets_upsert_keyed_record. Matches
	// SECRETS_TAG_UKEY in secrets.c.
	secretsTagUKey uint32 = 0x59454B55

	// secretsTagMID is the little-endian "M-ID" tag used by
	// secrets_ensure_mid_record for slot 0x7C. Matches
	// SECRETS_TAG_MID in secrets.c.
	secretsTagMID uint32 = 0x44492D4D
)

// OwnerPerms record constants — exact bytes the firmware emits in
// the synthesised default record (auth_derive_session_key fallback).
const (
	// ownerPermsPayload is the 16-byte application payload at +0x00.
	// The template string in the OEM binary is "F_____OWNER_PERMS"
	// (the leading 'F' comes from format-string deduplication); only
	// bytes [1..17) are copied into the record.
	ownerPermsPayload = "_____OWNER_PERMS"

	// ownerPermsKey is the lookup key written at payload+16. Because
	// it is zero, the synthesised record is the one returned for
	// `auth_derive_session_key(0)` on an unprovisioned device — and
	// in practice for any client_key_id on such a device, since the
	// fallback fires whenever find_by_key misses.
	ownerPermsKey uint32 = 0x00000000

	// ownerPermsPermissionMask is the 32-bit "permission mask" word
	// at +0x14, all bits set in the default record. Application-
	// level semantics; the secrets-store API itself does not inspect
	// this field.
	ownerPermsPermissionMask uint32 = 0xFFFFFFFF

	// ownerPermsMagic is the record-type tag "UKEY" at +0x18 stored
	// little-endian (firmware literal pool @ 0x00018BA8). Identical
	// to secretsTagUKey but kept separate for readability at the
	// OWNER_PERMS construction site.
	ownerPermsMagic uint32 = 0x59454B55
)

// ErrProvisioned is returned by DeriveOwnerPermsIfUnprovisioned when
// the dump shows the device has already been provisioned (mfg key
// present or at least one CRC-valid record in the keys region), so
// the firmware would not fall back to the default OWNER_PERMS key.
var ErrProvisioned = errors.New("device is provisioned — firmware would not synthesise default OWNER_PERMS")

// BuildOwnerPermsRecord returns the 32-byte default OWNER_PERMS
// session-key record exactly as the firmware materialises it inside
// auth_derive_session_key when the device is unprovisioned. The
// trailing 4 bytes hold the CRC-32/LE of the preceding 28 bytes, so
// the returned slice is a fully-formed secrets record that would
// round-trip through SecretsRecordRead / secrets_record_read.
func BuildOwnerPermsRecord() []byte {
	record := make([]byte, secretsRecordBytes)

	copy(record[0x00:secretsKeyOffset], []byte(ownerPermsPayload))
	binary.LittleEndian.PutUint32(record[secretsKeyOffset:secretsKeyOffset+4], ownerPermsKey)
	binary.LittleEndian.PutUint32(record[0x14:0x18], ownerPermsPermissionMask)
	binary.LittleEndian.PutUint32(record[secretsTagOffset:secretsTagOffset+4], ownerPermsMagic)

	crc := CalcCRC32LE(secretsCRCSeed, record[:secretsPayloadBytes])
	binary.LittleEndian.PutUint32(record[secretsPayloadBytes:], crc)

	return record
}

// SecretsRecordRead is the host-side equivalent of the firmware's
// `secrets_record_read` (@ 0x00020BB8). Given an SPI-flash dump and
// a slot index in [0, 127], it reads the 32-byte record at
//
//	secretsSectorBase + slot * secretsRecordBytes
//
// verifies the trailing CRC against the 28-byte payload, and on a
// match returns a copy of the full 32-byte record. Returns nil and
// no error when the slot is out of range or the CRC does not match
// (which is how the firmware signals "no valid record" — a missing
// record is not an exceptional condition). I/O errors are returned
// as-is.
func SecretsRecordRead(file *os.File, slot int) ([]byte, error) {
	if slot < 0 || slot > 0x7F {
		return nil, nil
	}

	offset := secretsSectorBase + int64(slot)*int64(secretsRecordBytes)
	buf := make([]byte, secretsRecordBytes)
	n, err := file.ReadAt(buf, offset)
	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("read slot %d at 0x%X: %w", slot, offset, err)
	}
	if n < secretsRecordBytes {
		return nil, nil
	}

	calc := CalcCRC32LE(secretsCRCSeed, buf[:secretsPayloadBytes])
	stored := binary.LittleEndian.Uint32(buf[secretsPayloadBytes:])
	if calc != stored {
		return nil, nil
	}
	return buf, nil
}

// SecretsCountValidInKeysRange is the host-side equivalent of the
// firmware's `secrets_count_valid_in_keys_range` (@ 0x00025680). It
// walks slots [0, 0x7B] and counts how many carry a CRC-valid
// record. Used by auth_derive_session_key to gate the default
// OWNER_PERMS fallback: any valid record in the keys region means
// the device has at least one provisioned client key, so the
// fallback must not fire.
func SecretsCountValidInKeysRange(file *os.File) (int, error) {
	count := 0
	for slot := secretsKeysRegionFirst; slot <= secretsKeysRegionLast; slot++ {
		rec, err := SecretsRecordRead(file, slot)
		if err != nil {
			return 0, err
		}
		if rec != nil {
			count++
		}
	}
	return count, nil
}

// IsUnprovisioned reports whether the SPI-flash dump matches the
// "untrusted" device state that auth_derive_session_key requires
// before it will synthesise the default OWNER_PERMS record:
//
//   - secrets_record_read(0x7E, NULL) must return 0 (no CRC-valid
//     manufacturing key at slot 0x7E), AND
//   - secrets_count_valid_in_keys_range() must return 0 (no
//     CRC-valid records in slots [0, 0x7B]).
//
// Both conditions are checked against the on-disk records; a dump
// produced from a factory-fresh module will satisfy them.
func IsUnprovisioned(file *os.File) (bool, error) {
	mfg, err := SecretsRecordRead(file, secretsMfgKeySlot)
	if err != nil {
		return false, err
	}
	if mfg != nil {
		return false, nil
	}

	n, err := SecretsCountValidInKeysRange(file)
	if err != nil {
		return false, err
	}
	return n == 0, nil
}

// DeriveOwnerPermsIfUnprovisioned ports the OWNER_PERMS-synthesis
// branch of auth_derive_session_key. It inspects the provided
// SPI-flash dump and:
//
//   - returns the 32-byte default OWNER_PERMS session-key record
//     (with CRC) when the device is in the un-provisioned state the
//     firmware requires (see IsUnprovisioned), or
//   - returns ErrProvisioned when the firmware would refuse the
//     fallback because a manufacturing key or at least one CRC-valid
//     key-region record is present.
//
// Use BuildOwnerPermsRecord directly if you want the synthesised
// bytes unconditionally (e.g. for offline crypto experiments) and do
// not care whether the live firmware would actually accept them.
func DeriveOwnerPermsIfUnprovisioned(file *os.File) ([]byte, error) {
	unprov, err := IsUnprovisioned(file)
	if err != nil {
		return nil, err
	}
	if !unprov {
		return nil, ErrProvisioned
	}
	return BuildOwnerPermsRecord(), nil
}

// SecretsFindByKey is the host-side port of firmware
// `secrets_find_by_key` (@ 0x00022BAA). It linear-scans the user-keyed
// region of the secrets sector (slots [0, 0x7B] inclusive) for a
// CRC-valid record whose payload+16 word equals key. On a match it
// returns the slot index and a copy of the full 32-byte record; on no
// match the slot index is -1 and the record is nil. Records with
// invalid CRCs are skipped, matching the firmware behaviour. I/O
// errors short-circuit the scan.
func SecretsFindByKey(file *os.File, key uint32) (int, []byte, error) {
	for slot := secretsKeysRegionFirst; slot <= secretsKeysRegionLast; slot++ {
		rec, err := SecretsRecordRead(file, slot)
		if err != nil {
			return -1, nil, err
		}
		if rec == nil {
			continue
		}
		slotKey := binary.LittleEndian.Uint32(rec[secretsKeyOffset : secretsKeyOffset+4])
		if slotKey == key {
			return slot, rec, nil
		}
	}
	return -1, nil, nil
}

// PrintPerms dumps every CRC-valid record in the user-keyed region
// of the secrets sector ([0, 0x7B]) and reports whether the firmware
// would synthesise the default OWNER_PERMS record on this dump.
//
// For each valid record the report is a multi-line block:
//
//   - slot index (dec/hex)
//   - the 16-byte application payload as hex (the BLE key material
//     for ordinary records, or "_____OWNER_PERMS" for the default)
//   - the same 16 bytes as printable ASCII when meaningful
//   - key id at payload+16 with a human-readable label when known
//     (BikeComm / Sharing / Workshop / OWNER_PERMS)
//   - permission mask at payload+20 (application semantics: which
//     bike functions this client key may invoke)
//   - record-type tag at payload+24 ("UKEY", "M-ID", ...)
//
// Then we probe the manufacturing-key slot (0x7E) and the
// IsUnprovisioned gate. If the dump is unprovisioned, the
// synthesised default OWNER_PERMS record is materialised and
// printed — that's the key the firmware will hand to any BLE client
// whose key id does not match an on-flash record.
func PrintPerms(file *os.File) error {
	fmt.Printf("Secrets sector @ 0x%X — user-keyed records (slots [0x%02X, 0x%02X]):\n\n",
		secretsSectorBase, secretsKeysRegionFirst, secretsKeysRegionLast)

	found := 0
	for slot := secretsKeysRegionFirst; slot <= secretsKeysRegionLast; slot++ {
		rec, err := SecretsRecordRead(file, slot)
		if err != nil {
			return err
		}
		if rec == nil {
			continue
		}
		found++
		printRecord(slot, rec)
	}
	if found == 0 {
		fmt.Println("  (no CRC-valid records found)")
	}

	mfg, err := SecretsRecordRead(file, secretsMfgKeySlot)
	if err != nil {
		return err
	}
	if mfg != nil {
		fmt.Printf("Manufacturing key (slot 0x%02X): PRESENT\n", secretsMfgKeySlot)
	} else {
		fmt.Printf("Manufacturing key (slot 0x%02X): absent\n", secretsMfgKeySlot)
	}

	unprov, err := IsUnprovisioned(file)
	if err != nil {
		return err
	}
	if !unprov {
		fmt.Println("Device state: PROVISIONED — firmware will NOT synthesise the default OWNER_PERMS record.")
		return nil
	}

	fmt.Println("Device state: UNPROVISIONED — firmware will synthesise the default OWNER_PERMS record:")
	printRecord(-1, BuildOwnerPermsRecord())
	return nil
}

// printRecord renders one 32-byte keyed record. Pass slot == -1 for
// the synthesised OWNER_PERMS record (skips the slot header line).
func printRecord(slot int, rec []byte) {
	keyMaterial := rec[:secretsKeyOffset]
	keyID := binary.LittleEndian.Uint32(rec[secretsKeyOffset : secretsKeyOffset+4])
	mask := binary.LittleEndian.Uint32(rec[0x14:0x18])
	tag := binary.LittleEndian.Uint32(rec[secretsTagOffset : secretsTagOffset+4])
	crc := binary.LittleEndian.Uint32(rec[secretsPayloadBytes:])

	if slot >= 0 {
		fmt.Printf("  slot %3d (0x%02X):\n", slot, slot)
	}
	fmt.Printf("    key id    : 0x%08X  (%s)\n", keyID, clientKeyName(keyID))
	fmt.Printf("    key (hex) : %s\n", strings.ToUpper(hex.EncodeToString(keyMaterial)))
	if ascii := printablePrefix(keyMaterial); ascii != "" {
		fmt.Printf("    key (ascii): %q\n", ascii)
	}
	fmt.Printf("    perm mask : 0x%08X\n", mask)
	fmt.Printf("    tag       : %q (0x%08X)\n", tagToASCII(tag), tag)
	fmt.Printf("    crc       : 0x%08X\n\n", crc)
}

// tagToASCII converts a little-endian 4-byte tag (as stored in the
// record at payload+24) into the ASCII string the firmware uses to
// identify it ("UKEY", "M-ID", ...).
func tagToASCII(tag uint32) string {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], tag)
	return string(b[:])
}

// printablePrefix returns the leading printable-ASCII run of buf,
// trimming trailing NULs and stopping at the first non-printable
// byte. Used so PrintPerms shows "_____OWNER_PERMS" rather than the
// raw 16-byte hex blob for human-meaningful payloads.
func printablePrefix(buf []byte) string {
	end := len(buf)
	for i, c := range buf {
		if c < 0x20 || c > 0x7E {
			end = i
			break
		}
	}
	return strings.TrimRight(string(buf[:end]), "\x00")
}

// DeriveSessionKey is the full host-side port of bleware's
// `auth_derive_session_key` (@ 0x00018B1C). Given an SPI-flash dump
// and a client-supplied 32-bit key id, it reproduces the firmware's
// resolution logic:
//
//  1. SecretsFindByKey(file, clientKeyID) — return the matching
//     CRC-valid record verbatim if one exists in slots [0, 0x7B].
//  2. Otherwise, *only* if the dump is unprovisioned
//     (IsUnprovisioned == true), synthesise the default OWNER_PERMS
//     record via BuildOwnerPermsRecord and return that.
//  3. Otherwise return ErrProvisioned — the firmware would return
//     NULL and refuse the session.
//
// The returned slice is always exactly secretsRecordBytes (32) on
// success and carries a valid trailing CRC, so it can be fed back
// into anything that consumes a secrets record.
func DeriveSessionKey(file *os.File, clientKeyID uint32) ([]byte, error) {
	_, rec, err := SecretsFindByKey(file, clientKeyID)
	if err != nil {
		return nil, err
	}
	if rec != nil {
		return rec, nil
	}
	return DeriveOwnerPermsIfUnprovisioned(file)
}
