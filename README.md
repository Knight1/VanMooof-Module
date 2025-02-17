# VanMooof-Module S3

This currently covers the MX25L51245GMI-08G-TR SPI Flash Chip on the SX3 and SX4 if you happen to have one of the few.

This Chip features 512Megabits (64 Megabytes) of Flash capacity.

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
- fmn key
- all Firmwares
    - Mainware (STM32F413VGT6 LQFP100)
    - Batteryware (STM32L072CZT6 LQFP48)
    - Shifterware 
    - Motorware
    - bleware (TMS320)
- Logs
- Shell (UART) for the Module, BLE, 

```
m2m.vanmoof.com
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

You need the backside of the PCB from the Module to dump the SPI Flash

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

Press ESC until the MCU reboots and holds itself in the Bootloader.

> [!CAUTION]
>Be careful to not delete the application with ea.
> 
>Otherwise you have a briked Cartdrige, as till now no Y-Modem Compatible Firmware file is available


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

Look for a HEX Value starting with 76 45 56 6A 47 46 and ending with 00 00 00
If you search for it in a Hex Editor you will notice the end very clearly because the next line begins with "Welcome to ES3"

the sha512sum of the Password is
7edd23b1c75e070db66475bb1869bee9dc64def2cb163dfea39ef8efcb534bf44db2da9e7307590222c516875fb4b07c7450556efd6520d986c5757ce3441bdd

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
enter the shell with `bledebug`
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

### Err 23


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
