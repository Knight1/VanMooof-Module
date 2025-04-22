# VanMooof-Module S3

This currently covers the MX25L51245GMI-08G-TR SPI Flash Chip on the SX3 and SX4 if you happen to have one of the few.

This Chip features 512Megabits (64 Megabytes) of Flash capacity.

<https://github.com/Omegaice/go-xmodem/blob/master/ymodem/ymodem.go>
<https://pkg.go.dev/github.com/sandacn/ymodem/ymodem>
<https://unix.stackexchange.com/questions/273178/file-transfer-using-ymodem-sz>

### Features

- ble keys (read/write)
  - bike comm
  - bike Sharing
  - Workshop
- bikecomm.vanmoof.com / ublox1.vanmoof.com
  - /upload
  - /ping-response
    - 'guid':'%s','statistics':{'batt':%d,'mac':'%s','swv':'%s','dist':%d}
  - /bike-message
- fmna key
  - fmna-rework (if you happen to have access to Apples FMNA API)
- all Firmwares
    - Mainware (STM32F413VGT6 LQFP100)
    - Batteryware (STM32L072CZT6 LQFP48)
    - Shifterware 
    - Motorware (F2806)
    - bleware (TMS320F28054F)
- Logs
- Shell (UART) for the Module, BLE, GSM


### Dead Module?

You can apply 12Vdc via the DC Plug and it only charges the Module. 

### Some Communication from the Bike to the VanMoof Backend via self-signed Certs.

The Backend only supports 
```aiignore
Hexcode  Cipher Suite Name (OpenSSL)       KeyExch.   Encryption  Bits     Cipher Suite Name (IANA/RFC)
 x3c     AES128-SHA256                     RSA        AES         128      TLS_RSA_WITH_AES_128_CBC_SHA256
```

The uuid is without dashes, 32chars, numbers and chars. No duplication checking.
dist is in Kilometers with hectometer. So 5,5 kilometers become 55 here.
responds with result true. 
```
curl -vk https://bikecomm.vanmoof.com/ping-response \
-H "Content-Type: application/json" \
-d '{"guid":"UUID","statistics":{"batt":95,"mac":"MAC","swv":"1.6.8","dist":37154}}'
```

Message Type(s) unknown
```
curl -vk https://bikecomm.vanmoof.com/bike-message 
-H "Content-Type: application/json" 
-d '{"mac_address":"MAC","message_type":"","message_data":""}'
```

/upload expects ublox Data. => InvalidUBloxDataException


## Known Firmware File SHA512 Sums

### Mainware
1.08.02 66cee63020ea35447fc7dcf41b61300715937f0f19d02ffdc1626ca0e8356fe00fff57fab0ef043077829129c4b66c40bd823a9e6ae0325c4d048227a5664587
1.09.03 52780d5fb984d954cc81a4ab2f72e612639b6573ea1c250bc96c0ee0707444a6fef7f82c6c3a347813ab14ca661116aef51a45eb9852ddf9cd53e83f16b35256

### Motorware
S.0.00.22 ce5815d55366a10decf724224d03a44d2d43ce3093de4d9f85a5ea646594b3cec59dd3a82227146f8069f41fcf381e341f391ccbf1f339c296cd768368a177cc

### Shifterware
0.237 8f454dfc1e600dfeae772465dd9791cde1b7588be22f7d88e88e61c9708634173a730be85ba19214d6c4544576ebeb8ea7e51e2686a163b77f7693292da97409

### BMSBoot
unknown cd2fdb29adc315da8b99d81d0ac18cacf13fbe0399a3763bc737df8b214fd6628804c1b55929da3d8a0f906ae8fc00884e108755152f6a840acfcb17460b3bcf

### BLEware
1.04.01 118084995f7423cf8b1c5589d49b20f203c06a4116213b4264a4c30d25060fee2fe057e1906e8a3ec9ab5323a02b2f72ee454fbc2c9cc7e6ca550ab71abcfe52
2.04.01 467f425f8ff329204876159697a71e04dec2b9fc7336892d233f68d7ce8ab8a4eb9b3dea506d5f885008a602301eb9a2ecbba66327379eb860115edd37a3057c

### Batteryware
1.14.1 dc3c1e3e731936f3c20dd6432e9a2b2855e7a699a179488e3d878438a2663d779d8640df2692e2008a885a0917fb270a95738ce75f731503ec70e7f3a6c72e02

### m2m.vanmoof.com (SMS the bike sends)
```
ALARM_BMS_REMOVED
SET_SHIPPING
START_FROM_SHIPPING
PLAY_FIRE
RIDING_MODE
CPU_STOP_MODE
CPU_STOPPED
SHOW_LOCK
AUTOWAKEUP
CARDRIDGE_REMOVED
LIPOCHARGE
RESET
OAD_FILE_TRF
OAD_UPDATE
DIAGNOSE
DIAG_RDY
OAD_FINISH
OAD_FAILED
OAD_RX_SOUND
FACTORY_TEST
PLAY_SHTDN
PLAY_LOCK_SHTDN
PLAY_LOCK_FROM_SLEEP
PLAY_SHTDN_RDY
ALARM_DELAY_ON
TURN_ON
LOW_SOC
PIN_START
PIN_STUCK
PIN_1ST
PIN_2ND
PIN_3ND
PIN_CHECK
PIN_OK
PIN_SHOW_OK
PIN_NOK
PIN_NOK_SHOW
UNLOCK
EXTRA_ALREADY_UNLOCKED
UNLOCK_COUNT
UNLOCK_COUNT_TIMEOUT
LOCK_PLAY_UNLOCK
LOCK_PLAY_START
LOCK_DIM_OFF
LOCK_CLEAR
LOCK_SETUP
LOCK_PIC
LOCK_COUNT
COUNT_OFF
COUNT_CLEAR
FIND_MY_PLAY
BIKE_SHIPPING_ACCIDENTAL_WAKE
BIKE_SHIPPING_LIPOCHARGE
POWERON
SMS_INIT
SMS_READ
SMS_WRITE
CTX_ACT
CTX_DEACT
PING_SEND
MESSAGE_SEND
LOCATION_SEND
POWEROFF
IDLE
WST_DISCHARGE_MODE
WST_CHARGE_MODE
WST_BYPASS_MODE
WST_NONE
```

### How to get started?

You need the backside of the PCB from the Module to dump the SPI Flash.

Tools needed: Torx Screw set. I used my iFixit Kit.

1. unlock bike and remove Module from the Frame
   2. if you do not unlock the bike, the Alarm stays on and will annoy you. I used duct tape to cover the speaker if i forgot it.
2. open Module and unscrew all internal screws of the PCB to remove the PCB. Make sure you unplug the Matrix LCD Cable carefully! You can replace the cable if you break it. 
3. On the backside of the PCB is the Macronix 16 Pin SPI Flash Chip near the port for the back light. 
4. Dump that Flash with an 16 Pin! SPI Flash Chip clamp and a Pi
   4.1 I used an Raspberry Pi Zero v1.1. There you have to enable the SPI Interface with raspi-config
5. If you screw it back to getter make sure to use Screw glue like Loctite. The (in)rush current from/to the battery and the ac voltage to the Motor is high especially if the battery has low charge thus low voltage. If a screw gets lose while you ride you create little sparks. 

```console
# sudo flashrom -p linux_spi:dev=/dev/spidev0.0 -r rom.rom
flashrom v1.2 on Linux 6.1.21+ (armv6l)
flashrom is free software, get the source code at https://flashrom.org

Using clock_gettime for delay loops (clk_id: 1, resolution: 1ns).
Using default 2000kHz clock. Use 'spispeed' parameter to override.
Found Macronix flash chip "MX66L51235F/MX25L51245G" (65536 kB, SPI) on linux_spi.
Reading flash... done.
```

### Bootloader

Press ESC on the UART Port until the MCU reboots and holds itself in the Bootloader. It will display.
```console
STM32 bootloader <1.09> Muco Technologies (c)2019
```


Example Bootloader Output:
```
'MT' (@) 2019 STM32F4, Stop
top

help
For more information on a specific command, type HELP command-name
help         This tekst
ver          Software version
reboot       reboot CPU
stm32        STM32 internal bootlaoder
crc          CRC32 check
ea           Erase application
es           Erase shadow app
em           Erase motor app
eb           Erase battery app
ec           Erase shifter app
ua           Upload app (Y-modem)
us           Upload shadow app (Y-modem)
um           Upload motor app (Y-modem)
ub           Upload battery app (Y-modem)
uc           Upload shifter app (Y-modem)
st           Start application
vi           Version information

ver
STM32 bootloader <1.09> Muco Technologies (c)2019

crc
APP CRC ok
No shadow application
MotorPcb CRC ok
Battery CRC ok
No Shifter application

vi
STM32 bootloader v1.09 (Feb 21 2020 14:50:53)
Loaded Application: v1.06.08 (Apr  9 2021 11:22:41) size 215108 bytes
No Shadow Application
No Shifter Application
Motor Application: v0.00.16 ( 03 2021 00:48:35) size 61720 bytes
Battery Application: v1.11.01 (Apr 19 2021 18:17:22) size 84020 bytes

ec
Erasing shifter flash 128 Kb... Erase sector 4
OK
```


pack-process�ÀFÀprocess pack files in external flash memory�source/monitor/cmd_packfs.c�Processing pakfs


### Shell login

Look for a HEX Value with ex. HxD starting with ```76 45 56 6A 47 46``` and ending with ```00 00 00```
If you search for it in a Hex Editor you will notice the end very clearly because the next line begins with "Welcome to ES3"

the sha512sum of the Password is
```
7edd23b1c75e070db66475bb1869bee9dc64def2cb163dfea39ef8efcb534bf44db2da9e7307590222c516875fb4b07c7450556efd6520d986c5757ce3441bdd
```

PBNjh0V46Eev8CcfS4LPJg

Commands available:
```
help
Available commands:
help              This tekst
reboot            reboot CPU
login             Login shell
logout            Logout shell
ver               Software version
distance          Manual set dst
gear              set gear
region            Region 0..3
blereset          hard reset BLE
bledebug          redirect uart8
show              Parameters
motorupdate       Update F2806 CPU
vollow            Audio volume
volmid            Audio volume
volhigh           Audio volume
wheelsize         Wheel 24/28 inch
speed             override speed
loop              main loop time
shipping          Shipping mode
logprn            Print log
logclr            Clear log 6
logapp            1/ 0
powerchange       1/ 0
factory           Load factory defaults
battery           Show battery
batware           Battery update
batreset          Battery reset
shiftware         Battery update
shifterstatus     Show shifter
shiftdebug        Show Modbus
shiftresetcounter   Reset shift counter
motorstatus       
gsminfo           Info from Ublox
gsmstart          start GSM function
gsmdebug          redirect uart2
bmsdebug          Show Modbus
sound             sample,volume,times
adc               read adc
bwritereg         Modbus Bat write register
bwritedata        Modbus Bat write data
breadreg          Modbus Bat read register
swritereg         Modbus Shift write register
swritedata        Modbus Shift write data
sreadreg          Modbus Shift read register
stc               read lipo monitor
stcreset          
setoad            test
setgear           save muco shifter
```
### bledebug Shell
enter the BLE Chip shell with `bledebug`
then execute reset to get this output:
```
bledebug
Connect to UART8
reset
Mon Feb 17 20:00:34 2025: Platform reset
Thu Dec 31 23:00:00 2020: This image is not provisioned


*** VANMOOF S3/X3 Monitor Program ***

BLE MAC Address: "74:d2:85:00:00:00"

Device name ................ : ES3-74D285000000
Firmware version ........... : 2.04.01
Compile date / time ........ : Mar 29 2021 / 14:17:30
Find-My accessory UUID ..... : N/A
Serialnumber ............... : N/A
BIM firmware version ....... : 1.00.01
BIM compile date / time .... : Jul 17 2020 / 14:53:15
reset type ................. : system reset
systick .................... : 705246
FMNA status ................ : not provisioned, deactivated
Time ....................... : Mon Feb 17 20:00:29 2025


Type 'help' for a list of all available commands.

```
Help output:

```
> help
The following commands are available:

    firmware-update                   - update a new image of firmware to the external flash
    extflash-verify                   - verify the current flashchip
    log-count                         - get log-count statistic
    log-dump <start-index> <n>        - print <n> blocks starting at address <start-index> 
    log-flush                         - flush all log-entries
    log-inject <n>                    - Create <n> fake-logs
    audio-play <index>                - play audio bound to the specified index
    audio-stop                        - stop playing the current audio file
    audio-dump                        - dump all audio files in external memory
    audio-upload <index>              - upload audio binary using Y-Modem at the address linked to the specified index
    audio-volume-set-all <level>      - set audio level of all audio-clips (0-3)
    pack-upload                       - upload a PACK file by Y-Modem
    pack-list                         - list the contents of a PACK file
    pack-delete                       - delete a PACK file
    pack-process                      - process pack files in external flash memory
    ble-info                          - dump current BLE connection info / statistics
    ble-disconnect                    - force a disconnect of all connected devices
    ble-erase-all-bonds               - erase all bonds
    fmna-get-auth-uuid                - Get the FMNA authentication UUID
    fmna-get-serialnumber             - Get the FMNA serialnumber
    fmna-set-serialnumber             - Set the serialnumber (hexstring, 8 bytes / 16 hexdigits)
    fmna-erase-external               - erases the external provisioned data (will be reset on reboot if internal exists)
    ble-erase-fmna-bonds              - will erase all fmna bonds
    fmna-blob-upload-int              - upload a FMNA factory blob
    fmna-rework                       - Replace provisioning data using Y-Modem
    fmna-enable-pairing               - enables pairing (if unpaired)
    fmna-unpair                       - Force FMNA unpair (i.e. 5x reset)
    shutdown                          - shutdown the system
    rtos-statistics                   - dump memory stats every 500ms
    rtos-nvm-compact                  - Compact the non-volatile storage
    reset                             - perform software reset of the MCU
    info/ver                          - show basic firmware info
    exit                              - exit from shell
    help                              - show all monitor commands

```

```
> pack-list
Scanning PACK archive...
 4,294,967,295 bytes ����������������������������������������������������������
 4,294,967,295 bytes ����������������������������������������������������������
 4,294,967,295 bytes ����������������������������������������������������������
 4,294,967,295 bytes ����������������������������������������������������������
 4,294,967,295 bytes ����������������������������������������������������������
```

```
> pack-process
Processing pakfs, expect a small startup delay because of mainware erasing its shadowflash, which blocks all serial I/O
Mon Feb 17 20:05:44 2025: Couldn't get new bleware image data

> Mon Feb 17 20:05:44 2025: Couldn't get new bleware image data
Mon Feb 17 20:05:45 2025: Couldn't get new bleware image data
Mon Feb 17 20:05:45 2025: Couldn't get new bleware image data
Mon Feb 17 20:05:45 2025: Couldn't get new bleware image data
Mon Feb 17 20:05:45 2025: Couldn't get new bleware image data
Wake Reason: WAKE_SRC_BLE
```

```
> log-flush
Done erasing logs
```

## Fixing some Errors
### Err 23

### Err missing DiSPlay

```console
I2C1 Error
01/00:07:02  ERR dsp freeze
```
you can not really login..

### Err Sim Card

The Module is looking for a SIM Card with a specific ICCID (Integrated Circuit Card Identifier).

The Prefix of that is 
```console
89 31 44 0400 
MM CC II N{12} C

MM = 89 (Mobile Networks)
CC = 31 (Netherlands, The)
II = 4X Vodafone
N{12} = Account ID
C = Checksum
```

### Firmware
We are looking for
```
50 41 43 4B BC 16 09 00 40 01 00 00 4F 41 44 20 (PACK¼	�@��OAD)
```

rom.rom

The OAD PACK is like, strings were in 108A5C
bleware.bin (It differs for the FMI Versions. Old Bikes do support FMI, it is just not in the Firmware. Just test it out.)
mainware.bin
motorware.bin
shifterware.bin
batteryware.bin


```
0x0002000 ?
0x005A000 BLE Secrets (60)
0x005af80 M-ID/M-KEY (60)
0x007c000 
0x0280000 VM_SOUND
0x0300000 VM_SOUND
0x0380000 VM_SOUND
0x2621440 VM_SOUND
0x3145728 VM_SOUND
0x4194304 VM_SOUND
0x4718592 VM_SOUND
0x5242880 VM_SOUND
0x5767168 VM_SOUND
0x6291456 VM_SOUND
0x640XXXX unknown
0x6815744 VM_SOUND
0x7340032 VM_SOUND
0x7864336 VM_SOUND
0x8388608 VM_SOUND
0x8912896 VM_SOUND
0x9437184 VM_SOUND
0x9961472 VM_SOUND
0x10485760 VM_SOUND
0x11010064 VM_SOUND
0x11534336 VM_SOUND
0x12058624 VM_SOUND
0x12582912 VM_SOUND
0x13107200 VM_SOUND
0x13631488 VM_SOUND
0x14155776 VM_SOUND
0x14680064 VM_SOUND
0x15204352 VM_SOUND
0x15728640 VM_SOUND
0x16252928 VM_SOUND
0x16777216 VM_SOUND
0x17301504 VM_SOUND
0x17825792 VM_SOUND
0x66965504 LOGS?
// WRONG?
0x3fdd000 LOGS


READ AND DECODE LOGS
6 GSM power ok
1723060256  Info ask tracking
1723060256 GSM_CMD_SMS_INIT
1723060256 GSM: sms= 0
1723060256 GSM_CMD_POWEROFF
1723060256 Poweroff g350..
1723060261 Poweroff g350 ok
1723060261 ES3.1 Main  1.08.02
1723060261 ES3 boot    1.09
1723060261 Motorware   S.0.00.22
1723060261 BMSWare     1.14 RSOC 100 Cycles 64 HW 3.10
1723060261 Shifterware 0.237 stored: 0.237
1723060261 BLEWare     2.4.01
1723060261 GSMWare     08.90
1723060261 CMD_BLE_MAC 24:9F:89:86:A9:1F
1723060261 PDOCP 0
1723060261 PDSCP 1
1723060261 iccid 89314404000979522399
1723060261 Modem ready
1723060261 GSM_CMD_IDLE
1723060280
```

### Important caveats.

This is all based on reverse Engineering. So there might be some Versions differences. 
So make a backup, save it in a save place like 1Password. If you compress the dump, the file gets very small.


### UART

#### Bootloader

For more information on a specific command, type HELP command-name
help         This tekst
ver          Software version
reboot       reboot CPU
stm32        STM32 internal bootlaoder
crc          CRC32 check
ea           Erase application
es           Erase shadow app
em           Erase motor app
eb           Erase battery app
ec           Erase shifter app
ua           Upload app (Y-modem)
us           Upload shadow app (Y-modem)
um           Upload motor app (Y-modem)
ub           Upload battery app (Y-modem)
uc           Upload shifter app (Y-modem)
st           Start application
vi           Version information
ver
STM32 bootloader <1.09> Muco Technologies (c)2019
ver
STM32 bootloader <1.09> Muco Technologies (c)2019
vi
STM32 bootloader v1.09 (Feb 21 2020 14:50:53)
Loaded Application: v1.08.02 (May  9 2022 10:58:01) size 220144 bytes
Shadow Application: v1.08.02 (May  9 2022 10:58:01) size 220144 bytes
Shifter Application: v0.ed.02 (Oct 23 2020 14:09:11) size 11944 bytes
Motor Application: v0.00.16 ( 03 2021 00:48:35) size 61720 bytes
No Battery Application
stm
'stm' is not recognized as an internal or external command.
crc
APP CRC ok
SHADOW CRC error
MotorPcb CRC ok
No Battery application
Shifter CRC ok



st
<Start application>
Wake Reason: WAKE_KICKLOCK

ES3 v1.08.02
NAK
NAK
NAK
ERR Led Display
ERR Light sensor
ERR ST3115 wake
01/00:07:01 SIM: PCB
RCC_FLAG_WWDGRST
RCC_FLAG_PINRST
01/00:07:01 GSM_CMD_IDLE
01/00:07:01 Set power state to PWR_NORMAL (Current limit: 20.0 A, SOC: -1 %)
NAK
NAK
NAK
NAK
NAK
NAK
01/00:07:01 BIKE_INIT
01/00:07:01 Restore power level 4
01/00:07:01 LiPo SoC 0% (first read)
01/00:07:01 Wake from shipping
I2C1 Error
01/00:07:01 BIKE_LIPOCHARGE
NAK
NAK
NAK
NAK
NAK
NAK
01/00:07:01  ERR dsp freeze
I2C1 Error
01/00:07:01  ERR dsp freeze
I2C1 Error
01/00:07:01  ERR dsp freeze
CMD_BLE_VERSION_INFO
CMD_BLE_MAC


#### Mainware


```console
Available commands:
help              This tekst
reboot            reboot CPU
login             Login shell
logout            Logout shell
ver               Software version
distance          Manual set dst
gear              set gear
region            Region 0..3
blereset          hard reset BLE
bledebug          redirect uart8
show              Parameters
motorupdate       Update F2806 CPU
vollow            Audio volume
volmid            Audio volume
volhigh           Audio volume
wheelsize         Wheel 24/28 inch
speed             override speed
loop              main loop time
shipping          Shipping mode
factory-shipping   Factory shipping mode (ignores BMS)
logprn            Print log
logclr            Clear log 6
logapp            1/ 0
powerchange       1/ 0
factory           Load factory defaults
battery           Show battery
batware           Battery update
batreset          Battery reset
shiftware         Battery update
shifterstatus     Show shifter
shiftdebug        Show Modbus
shiftresetcounter   Reset shift counter
motorstatus
gsminfo           Info from Ublox
gsmstart          start GSM function
gsmdebug          redirect uart2
bmsdebug          Show Modbus
sound             sample,volume,times
adc               read adc
bwritereg         Modbus Bat write register
bwritedata        Modbus Bat write data
breadreg          Modbus Bat read register
swritereg         Modbus Shift write register
swritedata        Modbus Shift write data
sreadreg          Modbus Shift read register
stc               read lipo monitor
stcreset
setoad            test
setgear           save muco shifter
```

```aiignore
ver
ES3.0 Main  1.08.02
ES3 boot    1.09
Motorware   BMSWare     0.00 RSOC 0 Cycles 0 HW 0.00
Shifterware 0.0 stored: 0.237
BLEWare     1.4.01
GSMWare
CMD_BLE_MAC F8:8A:5E:4F:9E:CB
```

```aiignore
show
 < I/O >
Botton_Left  0
Botton_Right 0
Lock         1
Wheel        0
Charger      1
Off/rst      0
MEMS_WAKE    0
BLE_WAKE     0
BAT_FAULT    1
BAT_KEY_IN   0
Fb coil det  0
Simcard      1
LiPo STAT    1
LiPo Status  LIPO_UNKNOWN
Serial       2043538395
Error Flags: 0x00000040 00000000
APP          DISCON
Ibat         20.0A
No extbat
 <Flash>
Dark         200 Lx now 591 Lx
backupcode   set
Volume Low    20
Volume Medium 30
Volume High   38
Group low     00000000
Group medium  383F33FE
Group high    47C0CC00
Unit         Metric
Wheel        28 inch
Transmission:AUTO
Light mode:  LIGHT_ALWAYS_ON
0 MaxSpeed 0 hm/h        0 %
1 MaxSpeed 160 hm/h      30 %
2 MaxSpeed 220 hm/h      30 %
3 MaxSpeed 270 hm/h      30 %
4 MaxSpeed 320 hm/h      30 %
5 MaxSpeed 320 hm/h      101 %
Region:      REGION_US
Moment 1 up:100  down:80 hm/h
Moment 2 up:190  down:170 hm/h
Moment 3 up:240  down:220 hm/h
Saved version: 10802F4
 <EEROM>
Alarm       Enable
log_by_app  Serial
Shipping    Active
Alarm state STANDBY 11
remote_locked 0
Horn file   10
Distance    4119.5 Km
Play_lock   1
Shifter         0.237
epoch_horn  0
shftr tries 0
 <GLOBAL>
pedal_speed: 0
torque     : 0
power level: 4
ride change: No
bike state   20
```

```aiignore

battery
BAT_ID  0x0
FAULT   0x0
BAT_TP1 -2731
BAT_TP2 -2731
MOS_TMP -2731
UBAT    0
RSOC    -1
I       0
CHARGE  0
DSG_PRT 0
TST_DBG 0
HW_VER  0.00
FW_VER  0.00
DATE_1  0
DATE_2  0
NOM_CAP 0
FUL_CAP 0
REM_CAP 0
ABS_SOC 0
CYL_CNT 0
CHG_ON  0
U_1 0
U_2 0
U_3 0
U_4 0
U_5 0
U_6 0
U_7 0
U_8 0
U_9 0
U_A 0
U_MAX 0
U_MIN 0
FSR   0x0000
DOTP  0
DUTP  0
COTP  0
CUTP  0
DOCP1 0
DOCP2 0
COCP1 0
COCP2 0
OVP1  0
OVP2  0
UVP1  0
UVP2  0
PDOCP 0
PDSCP 0
MOTP  0
SCP   0
```

### BLE DEBUG

```aiignore
Connect to UART8


*** VANMOOF S3/X3 Monitor Program ***

BLE MAC Address: "f8:8a:5e:4f:9e:cb"

Device name ................ : ES3-F88A5E4F9ECB
Firmware version ........... : 1.04.01
Compile date / time ........ : Mar 29 2021 / 14:20:30
BIM firmware version ....... : 1.00.00
BIM compile date / time .... : Apr 23 2020 / 14:10:12
reset type ................. : pin reset
systick .................... : 30316279

Type 'help' for a list of all available commands.
```

```aiignore

> help
The following commands are available:

    firmware-update                   - update a new image of firmware to the external flash
    extflash-verify                   - verify the current flashchip
    log-count                         - get log-count statistic
    log-dump <start-index> <n>        - print <n> blocks starting at address <start-index>
    log-flush                         - flush all log-entries
    log-inject <n>                    - Create <n> fake-logs
    audio-play <index>                - play audio bound to the specified index
    audio-stop                        - stop playing the current audio file
    audio-dump                        - dump all audio files in external memory
    audio-upload <index>              - upload audio binary using Y-Modem at the address linked to the specified index
    audio-volume-set-all <level>      - set audio level of all audio-clips (0-3)
    pack-upload                       - upload a PACK file by Y-Modem
    pack-list                         - list the contents of a PACK file
    pack-delete                       - delete a PACK file
    pack-process                      - process pack files in external flash memory
    ble-info                          - dump current BLE connection info / statistics
    ble-disconnect                    - force a disconnect of all connected devices
    ble-erase-all-bonds               - erase all bonds
    fmna-get-auth-uuid                - Get the FMNA authentication UUID
    fmna-get-serialnumber             - Get the FMNA serialnumber
    fmna-set-serialnumber             - Set the serialnumber (hexstring, 8 bytes / 16 hexdigits)
    fmna-erase-external               - erases the external provisioned data (will be reset on reboot if internal exists)
    ble-erase-fmna-bonds              - will erase all fmna bonds
    fmna-blob-upload-int              - upload a FMNA factory blob
    fmna-rework                       - Replace provisioning data using Y-Modem
    fmna-enable-pairing               - enables pairing (if unpaired)
    fmna-unpair                       - Force FMNA unpair (i.e. 5x reset
    shutdown                          - shutdown the system
    rtos-statistics                   - dump memory stats every 500ms
    rtos-nvm-compact                  - Compact the non-volatile storage
    reset                             - perform software reset of the MCU
    info/ver                          - show basic firmware info
    exit                              - exit from shell
    help                              - show all monitor commands
```

```
> extflash-verify
Flash type         : MX25L51245G
Device size (Mbyte): 64
SW status register write protection: y
Block write-protection enabled     : n
```

```aiignore
> pack-process
Sat Jan  1 00:55:04 2000: invalid pack content
Processing pakfs, expect a small startup delay because of mainware erasing its shadowflash, which blocks all serial I/O

> Cold boot

ES3 v1.08.02
```

```aiignore
> ble-info
number of connections: 0/3
Device address: cb:9e:4f:5e:94:f8
```

```aiignore
> ble-erase-all-bonds
```

```aiignore
> pack-list

```

```aiignore
> pack-delete
Deleting PACK archive...
Erase pack progress <99%>
Done
```

```aiignore
> pack-upload                                                                                                                                     
CCCCCCYModem successfully received 595968 bytes for file "pack.bin"
```

```
> pack-list
Scanning PACK archive..      
       217,884 bytes bleware.bin
       220,144 bytes mainware.bin
        61,720 bytes motorware.bin
        11,944 bytes shifterware.bin
        83,940 bytes batteryware.bin
```

```aiignore
> pack-process
Processing pakfs, expect a small startup delay because of mainware erasing its shadowflash, which blocks all serial I/O

> Sat Jan  1 02:18:53 2000: Platform reset
Thu Dec 31 23:00:00 2020: This image is not provisioned
Wake Reason: WAKE_KICKLOCK

ES3 v1.08.02
  ERR ST3115 wake
01/02:15:43 SIM: Holder
RCC_FLAG_PINRST
01/02:15:43 GSM_CMD_IDLE
01/02:15:43 Set power state to PWR_NORMAL (Current limit: 20.0 A, SOC: -1 %)
01/02:15:43 BIKE_INIT
01/02:15:43 Restore power level 4
01/02:15:43 LiPo SoC 0% (first read)
01/02:15:43 Wake from shipping
Seq 2 Order 1 2 3 4 5 
01/02:15:43  ERR dsp freeze
01/02:15:43 BIKE_OAD_UPDATE
CMD_BLE_VERSION_INFO
CMD_BLE_MAC
Disable Advertise
01/02:15:50 LiPo state changed to LIPO_ERROR
```
