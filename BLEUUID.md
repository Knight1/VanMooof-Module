# VanMooof ES3 + ES5 + S6 BLE UUID's

Every 128-bit UUID observed on VanMoof ES3 / ES5 / S6 modules — 172
distinct UUIDs in the latest dump. Characteristics that the firmware
declares but never exposes are tracked in
[Missing / unknown](#missing--unknown--consolidated). They
fall into a small number of *families* — each family shares a common
base and only a few hex digits change per entry. To make those
relationships visible, entries are grouped by base UUID and the
variable hex digits are tabulated separately rather than re-printing
the full 32-character UUID on every line.

Conventions used below:
- The fixed prefix and suffix of a family are shown once at the top
  of its section.
- `**XXXX**` (or `**XX**`) marks the bytes that vary; replace it with
  the listed short id to reconstruct the full UUID.
- Annotations cite the decompiled bleware **and** mainware sources
  where the UUID is consumed.

Source references:
- `bleware/src/gatt_read.c` — central GATT read dispatcher
- `bleware/src/gatt_write.c` — central GATT write dispatcher
- `bleware/src/monitor/cmd_audio.c` — `audio_play` / `audio_upload`
- `bleware/docs/progress.md` — service & characteristic registry
- `mainware/docs/ble-uuids.md` — UUID → 16-bit command id scheme
- `mainware/docs/ble-commands.md` — per-command-id action map
- `mainware/docs/state-machine.md` — lock/alarm states the commands drive

---

## How the scheme works

The UUID *table* physically lives in **bleware** (the CC2642 BLE
co-processor). mainware (STM32F413) never sees a 128-bit UUID — only
the 16-bit short id that bleware forwards over the inter-module bus.

### Two-layer model

VanMoof uses the standard vendor trick: **one 128-bit base UUID with a
16-bit "short" spliced into bytes `[2..3]`.**

```
6ACC0000-E631-4069-944D-B8CA7598AD50
    ^^^^
    bytes [2..3] = the 16-bit short id
```

This is exactly how the Bluetooth SIG embeds 16-bit UUIDs in the
standard base `0000xxxx-0000-1000-8000-00805F9B34FB` — VanMoof just
swapped in their own vendor base, so the whole attribute database is
"private" but addressed with two bytes internally.

### The `svc + idx + 1` rule

Service shorts sit on `0x__0` boundaries. Inside a service,
characteristics are numbered **sequentially from `svc + 1`**: the
0-based characteristic index `idx` has short id `svc + idx + 1`.

The key identity that makes the whole system legible:

> **the characteristic short id == the inter-module command id == the
> value mainware dispatches on.**

```
phone app                 bleware (CC2642)                 mainware (STM32F413)
---------                 ----------------                 --------------------
write 6ACC5521-…AD50  →   (svc 0x5520, idx 0)
       payload               cmd = 0x5520+0+1 = 0x5521  →  ble_cmd_dispatch case 0x5521
                             forward over SSP/Modbus bus      → lock state machine
notify 6ACC5521-…AD50 ←   pack response             ←       ssp_ble_enqueue_tx_packet(0x5521,…)
```

To go the other way — from a UUID you sniffed to what it does:
`svc = short & 0xFFF0`, `idx = (short & 0x000F) - 1`, then look up the
short in the per-service tables below.

The rule also cross-checks the characteristic counts in bleware's
service table: every count below reconstructs exactly to the short ids
observed on disk.

### The auth gate (why a raw write usually does nothing)

Two characteristics in the backoffice service are **not** plain bus
bridges:

- **`0x5502`** (`6ACC5502-…AD50`) — the **authentication handshake**.
  The app writes a 20-byte blob; bleware derives a 16-byte session key
  from a 4-byte seed inside it and stores it on the connection. Every
  other service's write callback first checks that a valid session key
  exists (per-characteristic "require session key" property bit).
  Without a successful `0x5502` handshake, writes to `0x5521`,
  `0x5523`, … are dropped *in bleware* and never reach mainware.
- **`0x5505`** (`6ACC5505-…AD50`) — the **backoffice** request/response
  channel (write *and* notify on the same characteristic), used for
  factory/secure operations rather than the simple command bridge.

For bench testing with nRF Connect / `bluetoothctl`: subscribe to the
characteristic you care about, complete the `0x5502` handshake first,
then write to e.g. `6ACC5521-…AD50` to drive lock/unlock.

### Where each half is implemented

| Concern | Firmware | Symbol / location |
| --- | --- | --- |
| 128-bit base + service/char tables | bleware | service table `0x0002A2F8` (11 entries `{u16 short_id, u16 char_count, u32 items_ptr}`) |
| Central write routing, `svc+idx+1` | bleware | `xs3_gatt_process_write_event` `0x00004DB0` |
| Auth handshake (`0x5502`) | bleware | `auth_derive_session_key` `0x00018B1C` |
| Command handlers (per short id) | mainware | `ble_cmd_dispatch` `0x08033970` |
| Read / telemetry handlers | mainware | `ble_read_request_dispatch` `0x08034D20` |

---

## Current VanMoof base — `6ACC` family

```
6ACC**XXXX**-E631-4069-944D-B8CA7598AD50
```

The canonical 128-bit base used by current bleware (`xs3_gatt_config.c`).
Bytes `[2..3]` of the UUID hold the short id — e.g. `6ACC5500-...` is
service `0x5500`. There are 11 services with their own characteristic
tables.

### Services (and characteristic count)

| Short id | # chars | Role |
|----------|---------|------|
| `5500` | 5 | **Backoffice / auth** — auth handshake + AES-decrypted request/response (factory + secure channel) |
| `5510` | 3 | **OAD** — firmware update (`oad_gatt_write_handler`, opaque handler) |
| `5520` | 4 | **Lock / alarm / power** — the anti-theft core |
| `5530` | 11 | **Ride config** — region, speed, gears, units, wheel (chars `5531..553B`) |
| `5540` | 18 | **Telemetry / status** — battery, motor, versions, errors; plus bleware-local ECC sign / TRNG |
| `5560` | 9 | **Alarm-arm / buttons / modem** + timekeeper (RTC) |
| `5570` | 4 | **Sound / backup code** — `cmd 0x5571` play forward to motor MCU, `cmd 0x5572` volume mask |
| `5580` | 4 | **Lights** — mode, LED channels, light sensor threshold |
| `5590` | 2 | (reserved / sparse) — typed publish bridge (`5591/5592`) |
| `55A0` | 4 | **Factory test** (`55A1..55A4`) |
| `55C0` | 3 | **Logs** — block count, total size, dump (opaque handler) |

> **Note:** the roles above come from the mainware command map. Several
> services that earlier bleware-side notes described generically as
> "Modbus bridge" are exactly that at the bleware layer — the bridge is
> the transport, and the table names what the forwarded command *does*
> once mainware receives it.

### Backoffice service `5500` — characteristics

| Char id | bleware role | mainware action |
|---------|--------------|-----------------|
| `5501` | Notify channel 0 | — |
| `5502` | **AUTH handshake** — 20-byte blob in, 4-byte seed → 16-byte session key | — (gates every other write) |
| `5503` | Notify channel 2 / `ssp_relay_u32(0x5503, u24)` | **W:** backup code / SoC — payload `FF FF FF` clears it (ctx+0x100 = 0xFFFF, alarm off); else ctx+0x100 = `p0 + 10·p1 + 100·p2`, alarm on; config-persist + ack. **R:** backup-code-set flag |
| `5504` | Notify channel 3 | — |
| `5505` | Backoffice request/response (240 B, AES-decrypted reassembly) | — (handled in bleware) |

### Service `5510` — OAD characteristics

| Char id | Role |
|---------|------|
| `5511` | OAD metadata |
| `5512` | OAD payload chunk |
| `5513` | (no-op in current dispatcher) |

### Service `5520` — lock / alarm / power

4 chars `5521..5524`. At the bleware layer these are Modbus bridges
(`module_forward_async` on write, idx 1 connection-state byte on read).

| Char id | mainware write | mainware read |
|---------|----------------|---------------|
| `5521` | **lock**: 1 = lock (state 0x20), 2 = unlock (reset buffers, state 0x30), 0 = if locked & lock-pin engaged → kick state 0x24 | **lock state** (`ble_lock_state_get`: 0 unlocked / 1 locked / 2 in PIN-lock) |
| `5522` | (bridge only) | (bridge only) |
| `5523` | **lock/unlock state machine**: 0 = unlock (state 0x0E), 1 = lock (state 0x0E), 2 = state 0x01, 3 = state 0x02, 4 = state 0x04; persists the state record | **unlock/alarm state** (`ble_unlock_state_get`, off alarm-mode ctx+0x310) |
| `5524` | **alarm on/off** (ctx+0x317); arms a display request, saves the state record, acks | alarm on/off (ctx+0x317) |

### Service `5530` — ride config

11 chars `5531..553B`. At the bleware layer: indices 2/4/5/7 are 1-byte
Modbus forwards; idx 3 is a 2-byte typed publish on cmd `5534`; idx 6 is
a 6-byte typed publish on cmd `5537`.

| Char id | mainware write | mainware read |
|---------|----------------|---------------|
| `5531` | **set odometer / distance** (ctx+0x31C) + ack | odometer / distance (ctx+0x31C) |
| `5532` | — | speed (ctx+0x3C2, scaled) |
| `5533` | **units** imperial/metric (ctx+0x10A) + config-persist + ack | units (ctx+0x10A) |
| `5534` | **motor power level / assist** (ctx+0x3C9 1..4); sets per-level ramp time ctx+0x354 (180/120/60/30) and the ride-change flag (ctx+0x3CB); `FUN_0803BEC8` | motor power level + ride-change (ctx+0x3C9/0x3CB) |
| `5535` | **region / speed mode** (ctx+0x109: 0=EU, 1=US, 2=JP, 3=OFFROAD); config-persist + display request + ack | region (ctx+0x109) + region lock (ctx+0x144) |
| `5536` | **shift gear** (ctx+0x3CC 1..4); increments the shift counter (ctx+0x338); `FUN_08028458` | shift gear (ctx+0x3CC) |
| `5537` | **speed "moments"** — six per-region up/down speed points (ctx + region·6 + 0x10E..0x12A, scaled ×10); config-persist | speed "moments" (per-region up/down, scaled) |
| `5538` | **transmission** auto/manual (ctx+0x108); config-persist | transmission (ctx+0x108) |
| `5539` | — | horn file (ctx+0x374) |
| `553A` | — | pedal speed (ctx+0x372) |
| `553B` | (bridge only) | (bridge only) |

> **Open discrepancy:** mainware also dispatches `0x553C` (region /
> speed mode *plus* the region lock, ctx+0x144), but bleware's `5530`
> table declares only 11 characteristics, so the highest reachable
> short id is `0x553B`. Either a different firmware revision widens the
> table, or `0x553C` is reachable only over the bus and not via GATT.
> Not resolved from the current dumps.

### Service `5540` — telemetry / status

18 chars, contiguous `5541..5552` by the `svc+idx+1` rule.

| Char id | mainware read | notes |
|---------|---------------|-------|
| `5541` | **battery summary** — SoC, temps (ctx+0x422/0x424), extbat (ctx+0x3D4/0x3DD/0x3DE/0x3E0/0x3E1) | |
| `5542` | BMS serial / version (ctx+0x3D5..0x3D7) | |
| `5543` | charger state (`FUN_08037160`) | |
| `5544` | LiPo status index (ctx+0x3D0) | |
| `5545` | — | not in the mainware read map |
| `5546` | motor telemetry (ctx+0x36C) | |
| `5547` | motor telemetry (ctx+0x36A) | |
| `5548` | **motor error flags** (ctx+0x3A4..0x3AB, 8 B) if any set | |
| `5549` | model / variant string (ctx+0x64A) | |
| `554A` | **app fw version** (image header @ `0x08020004`) — gated on state ≠ 3 | |
| `554B` | — | **bleware-local**: ECC sign with factory key (16 B), char idx 10 — never forwarded, which is why mainware has no case for it |
| `554C` | version string (ctx+0x388) | |
| `554D` | HW version (`hw_version_lookup`) | |
| `554E` | modem version (ctx+0x3E8) | |
| `554F` | shifter version (ctx+0x336) | |
| `5550` | powerbank version (ctx+0x408 + S/N) | |
| `5551` | testmode blob (`FUN_0802A9DC`, 0x60 B) | **conflict:** bleware notes char idx 16 as ECC sign with factory key (32 B). Both handlers claim this id — resolve before relying on it |
| `5552` | — | **bleware-local**: TRNG fill (16 B), char idx 17 |

Chars not claimed by mainware carry mfg-key-AES-encrypted static
identity (MAC, etc.) produced inside bleware.

### Service `5560` — alarm-arm / buttons / modem

9 chars `5561..5569`.

| Char id | mainware write | mainware read |
|---------|----------------|---------------|
| `5561` | — | coarse bike status (`bike_status_coarse_get`) |
| `5562` | **lock / alarm arm** — guarded state machine (checks bike state + lock-pin GPIO PC8): 1/2 → alarm-active (states 0x0E/0x07), 0 → lock or unlock-request (state 0x24 / 0x12) by pin state | alarm armed? (`bike_state_is_standby`, state 0x0E) |
| `5563` | — | **error / status flags** (ctx+0x3B8, 8 B — the Diag/Error Flags words, see ERRORS.md) |
| `5564` | **wheel size** (ctx+0x10B 24/28); toggles a GPIO + schedules a 1 s task; config-persist + ack | wheel size (ctx+0x10B) |
| `5565` | **modem restart** (`FUN_0803CE08`, zero the SMS-info step) if payload 1 | — |
| `5566` | **horn / sound select** (`FUN_08038EC4`, index < 7) + display request | — |
| `5567` | `FUN_080381D0(payload u32)` | tracking / modem field (`FUN_080380EC`) |
| `5568` | — | **button states** + lock pin (`FUN_08040350`/`FUN_08040368`, GPIOD PD2) |
| `5569` | — | supply OK? (`supply_voltage_read` ≥ 20000) |

At the bleware layer, idx 6 (`5567`) is the RTC: write
`timekeeper_submit_epoch` then `ssp_relay_u32(0x5567, timekeeper_read_be())`;
read produces the RTC upper word in big-endian. This is the same
characteristic mainware treats as the tracking/modem time field.

### Service `5570` — sound / backup code

| Char id | bleware role | mainware action |
|---------|--------------|-----------------|
| `5571` | `module_forward_async(0x5571, idx)` — audio play forward to motor | **W:** play notify/sound — `channel_notify_with_status(payload[0])` |
| `5572` | `module_publish_command(0x5572, payload, 12)` — per-channel volume mask | **W:** backup-code / group set (ctx+0xF4/0xF8/0xFC, the audio/group words; `FUN_08033914` = digit count) + config-persist + ack. **R:** backup-code group words |
| `5573` | **not present in any dump** — declared by the char count, never enumerated | — (no dispatch case) |
| `5574` | Also routes to `0x5571` on write | **W:** horn file select (ctx+0x318) + save record + ack. **R:** horn file (ctx+0x318) |

### Service `5580` — lights

| Char id | bleware role | mainware action |
|---------|--------------|-----------------|
| `5581` | 1-byte forward | **W:** light mode (ctx+0x10C: 0=auto, 1=on, 2=off); config-persist + display request. **R:** light mode |
| `5582` | 3-byte typed publish | **W:** LED control — three channels on/off via `FUN_0803C5FC`/`FUN_0803C5F0`/`FUN_0803C608`. **R:** LED states (`FUN_08037A68`×3) |
| `5583` | 1-byte forward (enumerated, no dispatch case) | — |
| `5584` | 2-byte `ssp_relay_u16` | **W:** dark/light threshold (ctx+0x102) + config-persist. **R:** light sensor (ctx+0x102) |

### Service `5590` — characteristics

`5591`, `5592` — typed publish bridge, declared by the char count but
**never enumerated in a dump** and with no mainware dispatch cases.
Reserved / sparse; purpose unknown.

### Service `55A0` — factory test

**None of these characteristics appear in a GATT dump** — the service
advertises but exposes nothing. Three of the four are nonetheless
implemented in mainware, so they are reachable over the inter-module bus
and presumably over GATT once factory/test mode is entered.

| Char id | In dump | mainware action |
|---------|---------|-----------------|
| `55A1` | no | set two flag bits (ctx+0xF0/0xF1) |
| `55A2` | no | **testmode config A** — build a hw-CRC descriptor {1, 10, 1, 0x100} and copy a 0x20-byte payload block |
| `55A3` | no | **testmode config B** — descriptor {0x0B, 9, 1, 0x100} + 0x20-byte payload block |
| `55A4` | no | **unknown** — no dispatch case either |

### Service `55C0` — log characteristics

| Char id | bleware role | mainware action |
|---------|--------------|-----------------|
| `55C1` | `log_block_count_get()` (4 BE bytes — 16-byte block count) | **W:** log-to-app toggle (ctx+0x313); payload 3 → request state 0x17 (factory/test). **R:** log-to-app (ctx+0x313), or 3 when state == 0x17 (DIAGNOSE) |
| `55C2` | `log_total_size_byte()` (effective total byte size = `(byte & 0xFFF) << 4`) | — |
| `55C3` | Log dump readout / control | — |

### Short ids observed in the latest GATT dump

All 11 services advertise: `5500`, `5510`, `5520`, `5530`, `5540`, `5560`,
`5570`, `5580`, `5590`, `55A0`, `55C0`.

Characteristics actually enumerated:
`5501..5505`, `5511..5513`, `5521..5524`, `5531..553B`, `5541..5552`,
`5561..5569`, `5571..5572`, `5574`, `5581..5584`, `55C1..55C3`.

Measured against the characteristic counts in bleware's service table,
**9 of the 67 declared characteristics never appear in the dump**:

| Missing short id | Declared by | Status |
|------------------|-------------|--------|
| `5573` | `5570` service (4 chars, only 3 enumerated) | **unknown** — no mainware dispatch case either; the only hole in an otherwise contiguous sound service |
| `5591`, `5592` | `5590` service (2 chars, 0 enumerated) | **unknown** — service advertises but exposes nothing; no mainware cases. Consistent with "reserved / sparse" |
| `55A1`, `55A2`, `55A3`, `55A4` | `55A0` service (4 chars, 0 enumerated) | **hidden** — `55A1`/`55A2`/`55A3` *are* implemented in mainware (`ble_cmd_dispatch`: flag bits, testmode config A/B) but the factory-test service exposes no characteristics in a normal dump. Likely gated behind factory/test mode. `55A4` has no mainware case |

Everything else in the `6ACC` family is complete and contiguous — including
`5537`, `5538` and `554F`, which earlier notes listed as scan gaps. They are
present; the earlier short-id list was wrong. Conversely `5573` was listed
as present and is not.

### Non-UUID command ids on the same dispatcher

`ble_cmd_dispatch` also handles low ids that are **not** GATT
characteristics — internal / provisioning traffic that arrives over the
inter-module bus only. Listed here so they are not mistaken for missing
UUIDs:

| cmd | action |
| --- | --- |
| `0x0A` | write ctx+0x388 (u32) |
| `0x0B` | write ctx+0x376 (u16) |
| `0x0C` | write motor params ctx+0x364/0x368/0x36C (u32) + 0x370 (u16) |
| `0x0D` | write ctx+0x372 (u32) |
| `0x14` | **read:** motor data block (ctx+0x378, Modbus poll to the motor) |
| `0x19` | **read:** battery data block (ctx+0x3B2, Modbus poll to the BMS) |
| `0xFA` | **power on/off**: payload[0]=1 — from shipping (state 3/4) → exit shipping (state 0x0B, save record, state 0x30); from state 5 → power on (state 0x0E + mode 0x0F) |
| `0x105` | bulk config write (memcpy 0x100 bytes → ctx+0x1DC, set valid flag ctx+0x2DC) |
| `0x10E` | `FUN_08029C0C(payload)` |
| `0x113` | write ctx+0x3E3 (u16); `FUN_0802A03C` |
| `0x118` | write ctx+0x38C (u32) |
| `0x119` | write ctx+0x3E2 (u8) |
| `0x11A` | **provision BLE MAC + serial** (ctx+0x390..0x395 MAC; `snprintf` serial → ctx+0x398) |
| `0x11B` | write ctx+0x37C/0x380/0x384 (u32) |
| `0x122` | `FUN_0802E3D0(u16)` |
| `0x123` | drive GPIO **PE5** (mask 0x20) high/low (BLE-chip reset line) |

Each `0x55xx` write is acked back to the phone with
`ssp_ble_enqueue_tx_packet(cmd, len, payload, 0)`.

---

## Legacy / parallel VanMoof base — `278D` family

```
278D**XXXX**-4692-039F-3445-A23FC55333D0
```

Mirrors the `6ACC` short-id layout — likely an older / parallel
advertisement so older Fixie app versions still see the bike. The
`svc+idx+1` rule and the command map above apply unchanged.

All 11 services advertise, same shorts as `6ACC`. Characteristics
enumerated:

`5501..5505`, `5511..5513`, `5521..5524`, `5531..5536`, `5539..553B`,
`5541..554E`, `5550..5552`, `5561..5569`, `5571..5572`, `5574`,
`5581..5584`, `55C1..55C3`.

**This family is narrower than `6ACC`** — it is missing three
characteristics that `6ACC` does enumerate, on top of the six gaps both
families share:

| Missing short id | Note |
|------------------|------|
| `5537` | speed "moments" — present in `6ACC`, absent here |
| `5538` | transmission auto/manual — present in `6ACC`, absent here |
| `554F` | shifter version — present in `6ACC`, absent here |
| `5573`, `5591`, `5592`, `55A1..55A4` | same gaps as `6ACC` |

That is consistent with `278D` being the older revision: the three
extras in `6ACC` are exactly the kind of settings added later (speed
points, transmission mode, shifter version reporting).

---

## Older VanMoof base — `6ACB` family

```
6ACB**XXXX**-E631-4069-944D-B8CA7598AD50
```

Same suffix as `6ACC`, base byte `[1]` is `B` instead of `C`. Suspected
earlier bike-model namespace. Fewer entries observed.

Short ids: `5501..5505`, `5507`, `5508`, `5511..5515`, `5522..5525`,
`5531..5533`.

Two structural differences from the other VanMoof families:

- **No service-level UUIDs are advertised at all** — no `5500`, `5510`,
  `5520`, `5530`. Only characteristic shorts appear, so the service
  grouping here is inferred, not observed.
- `5507`, `5508`, `5514`, `5515` and `5525` fall **outside** the
  characteristic counts of the current `6ACC` services. This family used
  a wider (or differently sized) table, so the current command map does
  not transfer id-for-id.

Missing / unknown in this family:

| Short id | Note |
|----------|------|
| `5506` | gap between `5505` and `5507` — unknown whether it ever existed |
| `5521` | **notable** — `5522..5525` are present but the first characteristic of the lock service is not. In `6ACC` this is the lock command |
| `5500`, `5510`, `5520`, `5530` | service UUIDs, never advertised (see above) |

---

## TI BLE-Stack base — `F000` family

```
F000**XXXX**-0451-4000-B000-000000000000
```

TI SimpleLink BLE5-Stack vendor base. Used by TI sample services and
the OAD stack.

| Short id | Role |
|----------|------|
| `5500` | Likely a TI service shadow |
| `FFC1`, `FFC2`, `FFC4` | TI OAD characteristics |
| `FFC3` | **missing** — in TI's stock OAD profile the four characteristics are `FFC1` identify / `FFC2` block / `FFC3` control-status / `FFC4` image-status. `FFC3` is absent from every dump, so either it is not instantiated in this build or the OAD control path is served through the `5510` service instead |

---

## VanMoof legacy — `8E7F1A` family

```
8E7F1A**XX**-087A-44C9-B292-A2C628FDD9AA
```

Only the last byte of the leading word varies (2 hex digits).

Short ids: `50`, `51`, `53`, `54`.

Missing / unknown: **`52`** — the sequence is otherwise contiguous
`50..54`, so a fifth entry very likely exists and simply was not
enumerated. Purpose of the family as a whole is still unidentified.

---

## VanMoof secondary legacy — `6567` family

```
6567000**X**-e001-1b19-1e17-103a23322331
```

Only one nibble varies.

Short ids: `1`, `2`, `3`.

---

## VanMoof unknown — `E3D8` family

```
E3D8000**X**-3416-4A54-B011-68D41FDCBFCF
```

Only one nibble varies.

Short ids: `0`, `1`.

---

## Bluetooth SIG 16-bit UUID

```
0000FF00-0000-1000-8000-00805F9B34FB
```

This is the standard Bluetooth SIG 128-bit form of the 16-bit UUID
`0xFF00` (a vendor-defined "user data" range). No family; this is the
only entry.

---

## Missing / unknown — consolidated

The latest dump contains **172 distinct UUIDs**. Everything below is
either declared-but-never-enumerated, or a hole in an otherwise
contiguous run. Nothing here has been observed on a bike — these are
the open ends.

| UUID | Family | Why it is expected | Status |
|------|--------|--------------------|--------|
| `6ACC5573-E631-4069-944D-B8CA7598AD50` | `6ACC` | `5570` declares 4 chars, 3 enumerated | unknown, no dispatch case |
| `6ACC5591-…AD50`, `6ACC5592-…AD50` | `6ACC` | `5590` declares 2 chars, 0 enumerated | unknown, no dispatch cases |
| `6ACC55A1-…AD50`, `6ACC55A2-…AD50`, `6ACC55A3-…AD50` | `6ACC` | `55A0` declares 4 chars, 0 enumerated | **implemented in mainware**, hidden from GATT — likely factory-mode gated |
| `6ACC55A4-…AD50` | `6ACC` | `55A0` char count | unknown, no dispatch case |
| `278D5537-…33D0`, `278D5538-…33D0`, `278D554F-…33D0` | `278D` | present in `6ACC`, absent here | older revision predates these |
| `278D5573`, `278D5591`, `278D5592`, `278D55A1..55A4` | `278D` | mirrors the `6ACC` gaps | same status as `6ACC` |
| `6ACB5506-…AD50` | `6ACB` | gap between `5505` and `5507` | unknown |
| `6ACB5521-…AD50` | `6ACB` | `5522..5525` present, first char absent | unknown |
| `6ACB5500/5510/5520/5530-…AD50` | `6ACB` | every other family advertises services | never advertised in this family |
| `F000FFC3-0451-4000-B000-000000000000` | `F000` | TI stock OAD profile has `FFC1..FFC4` | not instantiated, or served via `5510` |
| `8E7F1A52-087A-44C9-B292-A2C628FDD9AA` | `8E7F1A` | `50..54` otherwise contiguous | unknown |

Families with **no** apparent gaps: `6567` (`0001..0003`), `E3D8`
(`0000..0001`), and the SIG `0000FF00` entry.

The highest-value unknowns to chase are `55A1..55A3` (already decompiled
in mainware — just need the GATT-side trigger) and `6ACB5521`, which
would confirm whether the older namespace used the same lock command.
