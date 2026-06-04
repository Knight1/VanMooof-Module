# Find My (FMNA) factory blob

The FMI builds of `bleware` (TI CC2642R1F — e.g. `bleware 2.4.01`) carry Apple's
**Find My Network Accessory** reference stack (the firmware strings name the
upstream sources: `fmna_image_provisioning.c`, `fmna_nvm_platform.c`,
`fmna_crypto.c`, plus VanMoof's `xs3_fmna_image_provisioning.c`). The
accessory's Find My provisioning data — what the firmware calls the **"FMNA
factory blob"** — lives in the external SPI flash at **0x7C000**.

Non-FMI builds (`bleware 1.4.01` and earlier) have no FMNA code and leave this
region blank. Many S3/X3 bikes are wired for Find My but shipped without the FMI
firmware.

```
0x005A000  BLE Secrets (60)
0x005af80  M-ID/M-KEY (60)
0x007B000  FMNA swap sector            ┐ 16 KB FMNA NV region
0x007C000  FMNA factory blob (area A)  │ (0x7B000..0x7EFFF, erased as a unit
0x007D000  (FMNA / legacy v1)          │  by the firmware's fmna-erase-external)
0x007E000  (FMNA / legacy v1)          ┘
0x0280000  VM_SOUND ...
```

## On-flash record (area A, 0x7C000)

The firmware keeps the record in a two-area, wear-levelled store: **0x7C000** is
the live area (area A) and **0x7B000** is the swap sector used during an atomic
rewrite. The live record is:

| Offset | Size | Contents |
| --- | --- | --- |
| `0x7C000` | `0x530` (1328) | **AES-128-CBC ciphertext** (IV = 0), 83 blocks |
| `0x7C530` | `0x2B0` | `0xFF` erase padding |
| `0x7C7E0` | `0x20` | **SHA-256** of the first `0x7E0` bytes (`ciphertext ‖ padding`) |

The SHA-256 is verified before use; it covers the *ciphertext*, so it detects
flash corruption independently of the key.

## Encryption — device-bound

The AES-128 key is derived from the CC2642's **factory BLE MAC** (read from
`FCFG1 + 0x2E8` on the live chip):

```
key[i] = mac[i % 6] + i          (i = 0..15, byte-wise, mod 256)
```

where `mac` is the six MAC bytes in **FCFG1 little-endian order** — the reverse
of the printed `AA:BB:CC:DD:EE:FF` address. Because the key is a function of the
per-device MAC, a blob copied to another bike will not decrypt. (This is binding
/ tamper-evidence, not strong secrecy — the MAC is semi-public.)

There is also an **internal** master copy in the CC2642's own flash; the
firmware re-syncs 0x7C000 from it on boot (`sync_image_provisioning_area`), so
erasing the external area self-heals on the next reboot.

## Decrypted record (0x530 bytes)

| Offset | Size | Field |
| --- | --- | --- |
| `+0x000` | 1 | format version (`0x02`) |
| `+0x001` | 16 | accessory **serial number** (ASCII, NUL-padded) |
| `+0x011` | 8 | metadata / flags |
| `+0x019` | 65 | EC **P-256** public key (`0x04 ‖ X ‖ Y`) |
| `+0x05A` | 65 | EC **P-256** public key (`0x04 ‖ X ‖ Y`) |
| `+0x09B` | 16 | **software-authentication UUID** |
| `+0x0AB` | 1024 | **software-authentication token** (Apple, ASN.1/DER) |
| `+0x4AB` | 32 | key material |
| `+0x4CB` | 57 | EC **P-224** public key — Apple server key **Q_A** (`0x04 ‖ X ‖ Y`) |
| `+0x504` | 32 | key material |
| `+0x524` | 12 | unused (`0xFF`) |

The 1024-byte token is a signed ASN.1/DER structure (a SET of SEQUENCEs with an
embedded ECDSA signature). It is Apple-vended, one-time-use, and rotated on every
Find My pairing. Per Apple's spec, the token, its UUID, the serial number, and
the server public keys persist across a factory reset.

## Usage

```console
# Decode the blob from a dump (MAC auto-detected from the dump)
VanMooof-Module -f dump.rom -fmna

# Supply the BLE MAC explicitly (printed order); both byte orders are tried
VanMooof-Module -f dump.rom -fmna -fmna-mac 24:9F:89:86:A9:1F

# Also write the decrypted record and the raw 1024-byte token to disk
VanMooof-Module -f dump.rom -fmna -fmna-out bike
#   → bike.fmna.bin        (1328-byte decrypted record)
#   → bike.fmna-token.bin  (1024-byte software-auth token)
```

The MAC is resolved in this order: `-fmna-mac`, then the secrets-sector MAC at
`0x5AFE0` (the `…MOOF` record), then the `CMD_BLE_MAC` line in the boot log at
`0x3FDD000`. Decryption is self-validating: the tool accepts the result only when
the decrypted format-version byte is `0x02`, and it tries both MAC byte orders.

### Example

```console
$ VanMooof-Module -f dump.rom -fmna
Find My (FMNA) factory blob @ 0x07C000
  SHA-256 @ 0x07C7E0: OK ()
  BLE MAC: 24:9F:89:42:42:42 (boot log CMD_BLE_MAC)
  AES-128-CBC key: 16-Byte (derived from MAC, FCFG1/reversed order)
  format version ........ : 2
  serial number ......... : "ASY313371337"
  software-auth UUID .... : UUIDv4
  software-auth token ... : 1024 bytes, ASN.1/DER (head )
  ...
```

```console
# The token verifies as DER:
$ openssl asn1parse -inform DER -in bike.fmna-token.bin | head -3
    0:d=0  hl=3 l= 190 cons: SET
    3:d=1  hl=2 l=  78 cons: SEQUENCE
    5:d=2  hl=2 l=   1 prim: INTEGER           :01
```

## Firmware cross-reference (bleware 2.4.01)

Addresses are file offsets into `bleware_2.4.01.bin` (image base `0x0`, runtime
vector table at `0x90`):

| Function | Role |
| --- | --- |
| `0x24954` | load record: read `0x530` @area + `0x20` @area+`0x7E0`, SHA-256 verify, AES-CBC decrypt, check version == 2 |
| `0x25410` | AES-CBC decrypt in place (83 × 16 B, IV = 0) |
| `0x2a5c8` | derive AES key from FCFG1 MAC (`key[i] = mac[i%6] + i`) |
| `0x1bc14` | SHA-256 over `ciphertext ‖ 0xFF` up to `0x7E0` |
| `0x21640` | decrypting external-flash read primitive |
| `0x14f98` | NVM init: pick active area (0x7C000 / 0x7B000 swap), call loader |
| `0x2a014` | `fmna-erase-external`: erase 0x7B000/0x7C000/0x7D000/0x7E000 |
