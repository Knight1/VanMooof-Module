// fmna.go — decode the Find My Network Accessory (FMNA) "factory blob".
//
// The FMI builds of bleware (TI CC2642R1F, e.g. 2.4.01) carry Apple's Find My
// Network Accessory reference stack ("fmna-r5"). The accessory provisioning
// data is kept in the external SPI flash as a two-area, wear-levelled NV store:
//
//	0x7B000  swap sector (4 KB)
//	0x7C000  main sector (area A) — holds the live record
//
// The on-flash record at 0x7C000 is:
//
//	0x7C000  0x530 bytes  AES-128-CBC ciphertext (IV = 0)
//	0x7C530  0x2B0 bytes  0xFF erase padding
//	0x7C7E0  0x20  bytes  SHA-256 of the first 0x7E0 B (ciphertext || padding)
//
// The AES key is derived from the CC2642's factory BLE MAC (FCFG1 + 0x2E8):
//
//	key[i] = mac[i % 6] + i        (i = 0..15, byte-wise, mod 256)
//
// where mac is the six MAC bytes in FCFG1 little-endian order (the reverse of
// the printed AA:BB:CC:DD:EE:FF address). Decryption is therefore device-bound:
// a blob copied onto a different bike will not decrypt.
//
// Decrypted, the 0x530-byte record is the FMNA provisioning struct:
//
//	+0x000  1     format version (= 2)
//	+0x001  16    accessory serial number (ASCII, NUL-padded)
//	+0x011  8     metadata / flags
//	+0x019  65    EC P-256 public key   (0x04 || X || Y)
//	+0x05A  65    EC P-256 public key   (0x04 || X || Y)
//	+0x09B  16    software-authentication UUID
//	+0x0AB  1024  software-authentication token (Apple, ASN.1/DER)
//	+0x4AB  32    key material
//	+0x4CB  57    EC P-224 public key   (Apple server key Q_A, 0x04 || X || Y)
//	+0x504  32    key material
//	+0x524  12    unused (0xFF)
//
// See FMNA.md for the full derivation and the bleware reader cross-reference.

package vanmoof

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// External-flash geometry of the FMNA record (FMI bleware).
const (
	fmnaAreaBase   int64 = 0x0007C000 // main provisioning sector (area A)
	fmnaSwapBase   int64 = 0x0007B000 // swap sector
	fmnaBlobLen    int   = 0x530      // AES-CBC ciphertext == plaintext struct
	fmnaHashRegion int   = 0x7E0      // bytes the SHA-256 covers (blob + 0xFF pad)
	fmnaHashLen    int   = 0x20       // SHA-256 digest length
	fmnaVersion    byte  = 0x02       // expected decrypted format version
)

// Field offsets inside the decrypted 0x530-byte record.
const (
	fmnaOffVersion = 0x000
	fmnaOffSerial  = 0x001 // 16 B ASCII serial number
	fmnaOffMeta    = 0x011 // 8 B metadata / flags
	fmnaOffPubA    = 0x019 // 65 B P-256 public key (0x04||X||Y)
	fmnaOffPubB    = 0x05A // 65 B P-256 public key
	fmnaOffUUID    = 0x09B // 16 B software-auth UUID
	fmnaOffToken   = 0x0AB // 1024 B software-auth token (ASN.1/DER)
	fmnaTokenLen   = 0x400
	fmnaOffSecret1 = 0x4AB // 32 B key material
	fmnaOffQA      = 0x4CB // 57 B P-224 public key (Apple server key Q_A)
	fmnaQALen      = 57
	fmnaOffSecret2 = 0x504 // 32 B key material
)

// DumpFMNA reads the FMNA factory blob from a flash dump, verifies its
// SHA-256, decrypts it with the device-bound AES key, and prints the decoded
// provisioning fields. macOverride, if non-empty, supplies the BLE MAC used
// for key derivation (e.g. "24:9F:89:86:A9:1F"); otherwise the MAC is taken
// from the dump (secrets sector, then boot logs). If outPrefix is non-empty,
// the decrypted record and the 1024-byte software-auth token are written to
// "<outPrefix>.fmna.bin" and "<outPrefix>.fmna-token.bin".
func DumpFMNA(file *os.File, macOverride, outPrefix string) error {
	ciphertext := readFromFile(file, fmnaAreaBase, fmnaBlobLen)
	if len(ciphertext) < fmnaBlobLen {
		return fmt.Errorf("short read at 0x%X (%d bytes)", fmnaAreaBase, len(ciphertext))
	}
	storedHash := readFromFile(file, fmnaAreaBase+int64(fmnaHashRegion), fmnaHashLen)

	fmt.Printf("Find My (FMNA) factory blob @ 0x%06X\n", fmnaAreaBase)

	if allByte(ciphertext, 0xFF) || allByte(ciphertext, 0x00) {
		fmt.Printf("  ⚠ area is blank (%s) — this bike was never Find-My-provisioned\n",
			map[bool]string{true: "0xFF", false: "0x00"}[allByte(ciphertext, 0xFF)])
		return nil
	}

	// Integrity: SHA-256 over the first 0x7E0 B (ciphertext + 0xFF erase pad).
	calc := fmnaComputeHash(ciphertext)
	if subtle.ConstantTimeCompare(calc[:], storedHash) == 1 {
		fmt.Printf("  SHA-256 @ 0x%06X: OK (%s)\n", fmnaAreaBase+int64(fmnaHashRegion),
			strings.ToUpper(hex.EncodeToString(calc[:8]))+"…")
	} else {
		fmt.Printf("  SHA-256 @ 0x%06X: MISMATCH (stored %s… calc %s…)\n",
			fmnaAreaBase+int64(fmnaHashRegion),
			strings.ToUpper(hex.EncodeToString(storedHash[:min(8, len(storedHash))])),
			strings.ToUpper(hex.EncodeToString(calc[:8])))
	}

	// Resolve the BLE MAC, then decrypt. The FCFG1 byte order is the reverse
	// of the printed address, so try both orders and keep whichever yields a
	// valid (version == 2) record.
	mac, macSrc, err := resolveFMNAMac(file, macOverride)
	if err != nil {
		return err
	}
	fmt.Printf("  BLE MAC: %s (%s)\n", formatMAC(mac), macSrc)

	plain, used := fmnaTryDecrypt(ciphertext, mac)
	if plain == nil {
		return fmt.Errorf("decryption failed: version byte != 0x%02X for either MAC byte order "+
			"(wrong MAC? supply it with -fmna-mac AA:BB:CC:DD:EE:FF)", fmnaVersion)
	}
	fmt.Printf("  AES-128-CBC key: %s (derived from MAC, %s order)\n",
		strings.ToUpper(hex.EncodeToString(fmnaDeriveKey(used))), macOrderName(mac, used))

	printFMNAFields(plain)

	if outPrefix != "" {
		blobPath := outPrefix + ".fmna.bin"
		if err := os.WriteFile(blobPath, plain, 0o644); err != nil {
			return fmt.Errorf("write decrypted record: %w", err)
		}
		fmt.Printf("  → decrypted record written to %s (%d bytes)\n", blobPath, len(plain))

		tokPath := outPrefix + ".fmna-token.bin"
		token := plain[fmnaOffToken : fmnaOffToken+fmnaTokenLen]
		if err := os.WriteFile(tokPath, token, 0o644); err != nil {
			return fmt.Errorf("write token: %w", err)
		}
		fmt.Printf("  → software-auth token written to %s (%d bytes)\n", tokPath, len(token))
	}
	return nil
}

// printFMNAFields decodes and prints the provisioning struct fields.
func printFMNAFields(p []byte) {
	serial := strings.TrimRight(string(p[fmnaOffSerial:fmnaOffSerial+16]), "\x00")
	uuid := p[fmnaOffUUID : fmnaOffUUID+16]
	token := p[fmnaOffToken : fmnaOffToken+fmnaTokenLen]

	fmt.Printf("  format version ........ : %d\n", p[fmnaOffVersion])
	fmt.Printf("  serial number ......... : %q\n", serial)
	fmt.Printf("  software-auth UUID .... : %s\n", formatUUID(uuid))
	fmt.Printf("  software-auth token ... : %d bytes, %s (head %s…)\n",
		len(token), asn1Note(token), hex.EncodeToString(token[:12]))
	fmt.Printf("  metadata (+0x011, 8B) . : %s\n", hex.EncodeToString(p[fmnaOffMeta:fmnaOffMeta+8]))
	fmt.Printf("  EC P-256 key #1 (+0x019): %s%s\n", ecPrefix(p[fmnaOffPubA]), hex.EncodeToString(p[fmnaOffPubA:fmnaOffPubA+65]))
	fmt.Printf("  EC P-256 key #2 (+0x05A): %s%s\n", ecPrefix(p[fmnaOffPubB]), hex.EncodeToString(p[fmnaOffPubB:fmnaOffPubB+65]))
	fmt.Printf("  Apple Q_A P-224 (+0x4CB): %s%s\n", ecPrefix(p[fmnaOffQA]), hex.EncodeToString(p[fmnaOffQA:fmnaOffQA+fmnaQALen]))
	fmt.Printf("  key material   (+0x4AB) : %s\n", hex.EncodeToString(p[fmnaOffSecret1:fmnaOffSecret1+32]))
	fmt.Printf("  key material   (+0x504) : %s\n", hex.EncodeToString(p[fmnaOffSecret2:fmnaOffSecret2+32]))
}

// fmnaTryDecrypt decrypts the record with the key derived from mac, trying the
// given byte order and its reverse. It returns the plaintext and the MAC byte
// order that produced a valid record, or (nil, nil) if neither validates.
func fmnaTryDecrypt(ciphertext, mac []byte) (plain, usedMAC []byte) {
	for _, cand := range [][]byte{mac, reversed(mac)} {
		pt := fmnaDecrypt(ciphertext, fmnaDeriveKey(cand))
		if pt != nil && pt[fmnaOffVersion] == fmnaVersion {
			return pt, cand
		}
	}
	return nil, nil
}

// fmnaDeriveKey builds the 16-byte AES key from the 6 MAC bytes:
// key[i] = mac[i % 6] + i (mod 256). mac must be the FCFG1 little-endian order.
func fmnaDeriveKey(mac []byte) []byte {
	key := make([]byte, 16)
	for i := 0; i < 16; i++ {
		key[i] = mac[i%6] + byte(i)
	}
	return key
}

// fmnaDecrypt performs AES-128-CBC decryption with a zero IV. It returns nil if
// the key is the wrong size or the ciphertext is not block-aligned.
func fmnaDecrypt(ciphertext, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil || len(ciphertext)%aes.BlockSize != 0 {
		return nil
	}
	out := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, make([]byte, aes.BlockSize)).CryptBlocks(out, ciphertext)
	return out
}

// resolveFMNAMac determines the BLE MAC for key derivation. Precedence:
// explicit override, then the secrets-sector MAC (0x5AFE0, "…MOOF"), then the
// "CMD_BLE_MAC" line in the boot log. Returns the 6 MAC bytes in printed order
// (most-significant first); fmnaTryDecrypt handles the FCFG1 reversal.
func resolveFMNAMac(file *os.File, override string) (mac []byte, source string, err error) {
	if override != "" {
		mac = parseMAC(override)
		if mac == nil {
			return nil, "", fmt.Errorf("invalid -fmna-mac %q (want AA:BB:CC:DD:EE:FF or 12 hex chars)", override)
		}
		return mac, "-fmna-mac", nil
	}

	// Secrets sector: 12 ASCII hex chars followed by the "MOOF" signature.
	if buf := readFromFile(file, 0x0005AFE0, 16); len(buf) == 16 && bytes.HasSuffix(buf, []byte("MOOF")) {
		if mac = parseMAC(string(buf[:12])); mac != nil && !allByte(mac, 0x00) && !allByte(mac, 0xFF) {
			return mac, "secrets 0x5AFE0", nil
		}
	}

	// Boot log: "CMD_BLE_MAC 24:9F:89:86:A9:1F".
	if mac = scanLogForMAC(file); mac != nil {
		return mac, "boot log CMD_BLE_MAC", nil
	}

	return nil, "", fmt.Errorf("could not find a BLE MAC in the dump; supply it with -fmna-mac AA:BB:CC:DD:EE:FF")
}

var macLineRe = regexp.MustCompile(`CMD_BLE_MAC\s+([0-9A-Fa-f]{2}(?::[0-9A-Fa-f]{2}){5})`)

// scanLogForMAC searches the 128 KB circular log buffer at 0x03FDD000 for the
// first non-zero "CMD_BLE_MAC" address.
func scanLogForMAC(file *os.File) []byte {
	const logBase, logLen = 0x03FDD000, 0x20000
	buf := readFromFile(file, logBase, logLen)
	for _, m := range macLineRe.FindAllSubmatch(buf, -1) {
		if mac := parseMAC(string(m[1])); mac != nil && !allByte(mac, 0x00) && !allByte(mac, 0xFF) {
			return mac
		}
	}
	return nil
}

// parseMAC accepts "AA:BB:CC:DD:EE:FF", "AA-BB-…", or 12 bare hex chars and
// returns 6 bytes in printed order, or nil on error.
func parseMAC(s string) []byte {
	s = strings.NewReplacer(":", "", "-", "", " ", "").Replace(strings.TrimSpace(s))
	if len(s) != 12 {
		return nil
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil
	}
	return b
}

func formatMAC(mac []byte) string {
	parts := make([]string, len(mac))
	for i, b := range mac {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(parts, ":")
}

// macOrderName reports whether the working order matches the printed MAC or its
// reverse (the FCFG1 little-endian storage order).
func macOrderName(printed, used []byte) string {
	if bytes.Equal(printed, used) {
		return "printed"
	}
	return "FCFG1/reversed"
}

func formatUUID(u []byte) string {
	if len(u) != 16 {
		return hex.EncodeToString(u)
	}
	h := hex.EncodeToString(u)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}

// asn1Note gives a short hint about the software-auth token encoding.
func asn1Note(t []byte) string {
	if len(t) > 1 && (t[0] == 0x30 || t[0] == 0x31) {
		return "ASN.1/DER"
	}
	return "opaque"
}

func ecPrefix(b byte) string {
	if b == 0x04 {
		return "(uncompressed) "
	}
	return ""
}

// fmnaComputeHash returns the SHA-256 the firmware stores at area+0x7E0: the
// digest of the first 0x7E0 bytes of the sector, i.e. the ciphertext followed
// by 0xFF erase padding.
func fmnaComputeHash(ciphertext []byte) [32]byte {
	region := make([]byte, fmnaHashRegion)
	copy(region, ciphertext)
	for i := len(ciphertext); i < fmnaHashRegion; i++ {
		region[i] = 0xFF
	}
	return sha256.Sum256(region)
}

// FMNAProvisioned reports whether the dump carries a valid Find My (FMNA)
// factory blob at 0x7C000, verified by the record's own SHA-256. The digest
// covers the ciphertext, so this needs no AES key / BLE MAC. Used by -verify to
// account for the Find My NV region (most bikes are not Find-My-provisioned).
func FMNAProvisioned(data []byte) bool {
	base := int(fmnaAreaBase)
	if len(data) < base+fmnaHashRegion+fmnaHashLen {
		return false
	}
	ct := data[base : base+fmnaBlobLen]
	if allByte(ct, 0xFF) || allByte(ct, 0x00) {
		return false
	}
	calc := fmnaComputeHash(ct)
	stored := data[base+fmnaHashRegion : base+fmnaHashRegion+fmnaHashLen]
	return subtle.ConstantTimeCompare(calc[:], stored) == 1
}

func allByte(b []byte, v byte) bool {
	for _, x := range b {
		if x != v {
			return false
		}
	}
	return len(b) > 0
}

func reversed(b []byte) []byte {
	r := make([]byte, len(b))
	for i := range b {
		r[len(b)-1-i] = b[i]
	}
	return r
}
