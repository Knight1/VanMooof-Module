// bleboot.go — host-side port of the BIM OAD slot scanner from
// bleboot/src/oad.c (`bim_full_scan_and_launch`, `bim_slot_iterator`,
// `oad_magic_match`).
//
// On reset BLEBoot probes the external SPI flash for an OAD image to
// promote into the CC2642's internal flash. It walks 44 candidate
// slots (0..43) at a 4 KB stride — i.e. offsets 0x00000, 0x01000,
// 0x02000, ... 0x2B000 — sniffing the 8-byte "OAD NVM1" magic at the
// start of each slot. The famous "BLEBoot tries to boot from 0x2000"
// is just slot 2 in that walk; the firmware accepts any of the 44.
//
// For every slot whose first 8 bytes match the magic, the BIM then
// parses a 44-byte short OAD header (`oad_short_header_t` in oad.c)
// and checks four acceptance gates before considering the slot a
// launch candidate:
//
//   hdr[12] == 0x03   (magic_a — BIM major)
//   hdr[13] == 0x01   (magic_b — BIM minor)
//   hdr[16] == 0xFE   (magic_c — full-scan view)
//   hdr[17] != 0xFC   (status not "rejected")
//
// A slot that passes all four is CRC-verified against hdr[8] (the
// stored CRC) over bytes [12, 12+image_size) — that CRC routine is
// `bim_crc32_buffer` and uses a non-standard seed/poly we have not
// finished mapping, so this host-side validator stops at the byte-
// pattern checks. The remaining metadata (image_size at +24,
// entry at +28, flags at +18, softVer at +32) is still printed so
// you can cross-check against the OEM bleware.bin.

package vanmoof

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	// BLEBootSlotStride is the byte stride between consecutive OAD
	// staging slots on the SPI flash. Matches `slot << 12` in
	// bim_slot_iterator.
	BLEBootSlotStride int64 = 0x1000

	// BLEBootSlotCount is the number of slots BLEBoot probes
	// (0..43 inclusive). The iterator stops at index 44.
	BLEBootSlotCount int = 44

	// BLEBootHeaderSize is the size of the OAD short header read by
	// the BIM scan path. Matches `oad_short_header_t` in oad.c.
	BLEBootHeaderSize int = 44

	// Legacy alias retained for any caller that still references the
	// "image lives at 0x2000" assumption — points at slot 2.
	BLEBootImageOffset int64 = 0x2000
)

// bleBootMagic is the 8-byte image-ID tag at offset 0 of every
// legitimate OAD header. Matches `OAD_MAGIC_A` / `OAD_MAGIC_B` in
// oad.c (the OEM keeps two identical copies in .rodata).
var bleBootMagic = [8]byte{'O', 'A', 'D', ' ', 'N', 'V', 'M', '1'}

// BLEBootImage is the decoded OAD short header for one staging slot.
// Field offsets mirror `oad_short_header_t` in oad.c; only fields the
// BIM actually consults during the scan are exposed.
type BLEBootImage struct {
	Slot      int
	Offset    int64
	CRC32     uint32   // +0x08 — expected CRC32 over [12, 12+ImageSize)
	MagicA    uint8    // +0x0C — must be 3
	MagicB    uint8    // +0x0D — must be 1
	MagicC    uint8    // +0x10 — must be 0xFE
	Status    uint8    // +0x11 — 0xFE verified, 0xFF pristine, 0xFC rejected
	Flags     uint8    // +0x12 — must be in {0, 1, 3, 7} for launch
	ImageSize uint32   // +0x18 — total image length
	Entry     uint32   // +0x1C — branch target
	SoftVer   [4]byte  // +0x20 — firmware version, BCD-like
}

// SoftVerString renders SoftVer the way the OEM console does — e.g.
// {0x00, 0x01, 0x04, 0x01} → "1.04.01" — by skipping the leading
// byte (always 0 in practice) and dotting the remaining three.
func (h BLEBootImage) SoftVerString() string {
	return fmt.Sprintf("%d.%02d.%02d", h.SoftVer[1], h.SoftVer[2], h.SoftVer[3])
}

// Acceptable reports whether the slot would pass the BIM's
// byte-pattern gates in bim_full_scan_and_launch. A return of true
// does NOT mean the CRC validates — that's a separate step the BIM
// runs next, which this host-side tool does not yet reproduce.
func (h BLEBootImage) Acceptable() bool {
	return h.MagicA == 0x03 &&
		h.MagicB == 0x01 &&
		h.MagicC == 0xFE &&
		h.Status != 0xFC
}

// PromotionState describes the persistent state the BIM has
// recorded about this slot via the +16 and +17 byte writes. The
// firmware updates these markers on every boot when it processes
// a slot — steps 4 (transient 0xFC to +16), 6 (0xFE to +17 on CRC
// match) and 7 (0xFC to +17 on CRC mismatch) of
// bim_full_scan_and_launch. They survive reboots and are how the
// BIM avoids re-promoting an image it already validated, and how
// it remembers to skip an image that has previously failed CRC.
//
// Returned values:
//
//   "pristine — never validated by BIM"
//   "in-progress — BIM started promotion but did not complete"
//   "verified — CRC matched on a previous boot, BIM will boot it"
//   "rejected — CRC failed on a previous boot, BIM permanently skips"
//   "unexpected magic_c=0xXX status=0xYY" for anything else.
func (h BLEBootImage) PromotionState() string {
	switch {
	case h.MagicC == 0xFE && h.Status == 0xFF:
		return "pristine — never validated by BIM"
	case h.MagicC == 0xFC && h.Status == 0xFF:
		return "in-progress — BIM started promotion but did not complete"
	case h.Status == 0xFE:
		return "verified — CRC matched on a previous boot, BIM will boot it"
	case h.Status == 0xFC:
		return "rejected — CRC failed on a previous boot, BIM permanently skips"
	}
	return fmt.Sprintf("unexpected magic_c=0x%02X status=0x%02X", h.MagicC, h.Status)
}

// LaunchableFlags reports whether the `flags` byte at +0x12 matches
// the launch-allowed set from bim_full_scan_and_launch
// ({0, 1, 3, 7}). The BIM consults this only after CRC validation.
func (h BLEBootImage) LaunchableFlags() bool {
	switch h.Flags {
	case 0, 1, 3, 7:
		return true
	}
	return false
}

// ScanBLEBootSlots reads every candidate slot the BIM walks and
// returns the parsed short header for each slot whose first 8 bytes
// match the "OAD NVM1" magic. Slots without the magic — the common
// case on a normal dump where every slot is 0xFF — are skipped
// silently. I/O errors are returned as-is.
func ScanBLEBootSlots(file *os.File) ([]BLEBootImage, error) {
	var out []BLEBootImage
	buf := make([]byte, BLEBootHeaderSize)

	for slot := 0; slot < BLEBootSlotCount; slot++ {
		offset := int64(slot) * BLEBootSlotStride
		n, err := file.ReadAt(buf, offset)
		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("read slot %d at 0x%X: %w", slot, offset, err)
		}
		if n < BLEBootHeaderSize {
			break
		}

		var magic [8]byte
		copy(magic[:], buf[:8])
		if magic != bleBootMagic {
			continue
		}

		img := BLEBootImage{
			Slot:      slot,
			Offset:    offset,
			CRC32:     binary.LittleEndian.Uint32(buf[0x08:0x0C]),
			MagicA:    buf[0x0C],
			MagicB:    buf[0x0D],
			MagicC:    buf[0x10],
			Status:    buf[0x11],
			Flags:     buf[0x12],
			ImageSize: binary.LittleEndian.Uint32(buf[0x18:0x1C]),
			Entry:     binary.LittleEndian.Uint32(buf[0x1C:0x20]),
		}
		copy(img.SoftVer[:], buf[0x20:0x24])
		out = append(out, img)
	}
	return out, nil
}

// PrintBLEBootImage writes a status line per OAD slot. The
// firmware's bim_full_scan_and_launch picks the first slot that
// passes all gates AND CRC-validates; we surface every slot with a
// magic hit so partially-written or rejected images are visible too.
func PrintBLEBootImage(file *os.File, fileSize int) {
	slots, err := ScanBLEBootSlots(file)
	if err != nil {
		fmt.Printf("BLEBoot OAD scan: error: %v\n", err)
		return
	}
	if len(slots) == 0 {
		fmt.Printf("BLEBoot OAD scan: no \"OAD NVM1\" magic in slots 0..%d (0x0000..0x%X) — BIM would not launch\n",
			BLEBootSlotCount-1, (BLEBootSlotCount-1)*int(BLEBootSlotStride))
		return
	}

	fmt.Printf("BLEBoot OAD scan: %d candidate slot(s) with \"OAD NVM1\" magic\n", len(slots))
	for _, img := range slots {
		verdict := bleBootVerdict(img, fileSize)
		fmt.Printf("  slot %2d @ 0x%05X: SoftVer %s  size=%d (0x%X)  entry=0x%08X  crc=0x%08X\n",
			img.Slot, img.Offset, img.SoftVerString(),
			img.ImageSize, img.ImageSize, img.Entry, img.CRC32)
		fmt.Printf("    magic_a=0x%02X magic_b=0x%02X magic_c=0x%02X status=0x%02X flags=0x%02X  → %s\n",
			img.MagicA, img.MagicB, img.MagicC, img.Status, img.Flags, verdict)
		fmt.Printf("    BIM state : %s\n", img.PromotionState())
	}
}

// bleBootVerdict mirrors the gates in bim_full_scan_and_launch and
// summarises what the BIM would do with this slot — short of the
// CRC step, which is deferred.
func bleBootVerdict(img BLEBootImage, fileSize int) string {
	switch {
	case img.Status == 0xFC:
		return "REJECTED (status=0xFC) — BIM skips"
	case img.MagicA != 0x03:
		return fmt.Sprintf("invalid (magic_a=0x%02X, expected 0x03)", img.MagicA)
	case img.MagicB != 0x01:
		return fmt.Sprintf("invalid (magic_b=0x%02X, expected 0x01)", img.MagicB)
	case img.MagicC != 0xFE:
		return fmt.Sprintf("invalid (magic_c=0x%02X, expected 0xFE)", img.MagicC)
	case img.ImageSize == 0, img.ImageSize == 0xFFFFFFFF:
		return "header present but image_size is blank — likely never written"
	case img.Offset+int64(img.ImageSize) > int64(fileSize):
		return fmt.Sprintf("TRUNCATED (declared size 0x%X overruns dump)", img.ImageSize)
	case !img.LaunchableFlags():
		return fmt.Sprintf("byte gates pass, but flags=0x%02X not in {0,1,3,7} (quick-scan would skip)", img.Flags)
	}
	return "byte gates pass — BIM would CRC-verify and launch on match"
}
