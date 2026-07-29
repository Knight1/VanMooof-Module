# VanMoof S3/X3 Error Repairs & Fixes

*Detailed repair procedures and troubleshooting steps for VanMoof bike error codes*

## 🔋 Battery & Power Repairs (Errors 0-21)

The first thing I recommend is setting the lights to always on if the module still has power.
When no light comes on, the BMS is not outputting any voltage.

### Critical Battery Errors

#### Main Errors 0, 6, 9, 10, 16

Most likely also displayed 17, 20

Fuse SFK40-45 likely an open line.  
Check fuse. Check pack and cell voltages. Check for visual damage & burns. Reset BMS.  
More in the dedicated BMS Repo. Ask me. Not for everyone.  

### Voltage Problems (Battery and bike dead)

#### Errors 4, 5
When the module is dead, you will not see them.  

Battery is likely sitting at 4.2V pack voltage. You need to charge the battery manually.  
Also, the module battery is likely empty.  

### Temperature Problems

#### Errors 12 - 15

Check temperature sensors in the battery if the problem persists in good temperature conditions.  

Check manual for resistance values and please add them.  

### System Communication

#### Error 19

There is a chip on the module which translates the BMS and the eShifter communication.  
When this chip goes bad, it might show this error.  

Another problem can be that the STM32 on the BMS does enable output, but is not communicating or cannot read cell voltages.  
You need to check the capacitors going from the cell monitor lines.  

Another problem can be that the STM32 on the BMS bootloops. In this case, you will not get any power.  

### Charging & Current Issues

#### Error 21

No current detected while in charing state.  

Potential impact: Bike no longer charges  
Set condition: Charger detected but no current  
Clear condition: Charger detected with current  

Make sure the charger is outputting 42 volts and the battery is not at 80% if a charge limit is set.  
The charger on the S3 does not check for any conditions to be met to supply voltage.  

## 📡 Communication & System Repairs (Errors 22-44)

### Communication

#### Error 22

No communication with the motor controller (TMS320F28054F).  
The module sends telemetry requests to the motor over the internal module bus and gets no
answer back. Every message is retried, and once more than four go unanswered the bike raises 22.

Potential impact: no assist. Motor temperature and motor firmware version stop updating.
Set condition: more than four messages to the motor left unacknowledged.  
Clear condition: any valid reply from the motor - it clears itself immediately.  

The bike tries to fix this on its own: while parked it pulses the motor reset line and
re-announces. On the debug console you will see `Resetting motor` followed by `Recover Motor ok`
if that worked. A bike that loops on `Resetting motor` has a motor controller that is not
coming back.

Check with `motorstatus` on the console; `Motor no resp` means nothing is answering at all.
Possible causes:
- Motorware update interrupted or corrupted (if you tried that)
- TMS320F28054F dead or firmware broken -> JTAG2
- No voltage input

#### Error 23

The bike shows "Err 23" and the app is unable to connect to the bike.  
In the meantime, you can still unlock the bike via the backup code.  

Potential impact: No communication with the bike via the app possible.  
Set condition: The STM32 chip is unable to communicate with the Texas Instruments CC2642 chip and resets the chip.  
Clear condition: The STM32 chip is able to communicate again.  
Possible causes:  
- Firmware corrupted  
- Chip dead  
- No voltage input  

This explains how to flash new firmware.  
If your chip is dead, you need to solder a new one onto the PCB with a hot air station, change the secondary MAC and change the bleware to use it.
  
1. Getting all the tools you need  
    a. XDS110 adapter from your friendly Chinese supplier  
    b. Torx screwdrivers  
    c. For now, a Windows PC  
    d. A Texas Instruments account to download the software  
    e. Possibly a USB extension cord  
    f. Loctite  
    g. 12Vdc permanently attached to the charging port of the PCB (or another method to supply all power rails of the CC26xx chip)  

For the clip:  
    h. Test stand PCB clip with 2x5 pins  


For soldering:  
    h. Soldering iron with small tip  
    i. Small wires  
    
  
2. Setting up the Hardware
  
You have to open the module completely and remove the PCB from the housing.  
On the underside of the PCB you will find the modem with the big white sticker on it. Next to it is a 2x5 gold-plated pin array with 10 pins.  
This is labeled JTAG1 with a marking 1 and 10. This marks the pin numbers.  
There is also JTAG2 on the PCB which is the wrong one.  
  
##### JTAG1 Pin Layout

Underside of the PCB, module held in the orientation it sits in the frame:
front of the bike to the **left**, rear to the **right**. In this view the  
charging plug is to the left.  

```
                   FRONT of bike                REAR of bike
                 (frontlight side)            (rearlight side)

                     +-------------------------------+
 *   RESET_N       --|  10 o                   o  5  |-- GND
     CC2642 pin 27 --|   9 o                   o  4  |-- unknown
     CC2642 pin 26 --|   8 o                   o  3  |-- GND
 *   JTAG_TCKC     --|   7 o                   o  2  |-- GND                   *
 *   JTAG_TMSC     --|   6 o                   o  1  |-- VTref / VDDS (3.3 V)  *
                     +-------------------------------+

 *  = the five pins you actually wire to the XDS110 (1, 2, 6, 7, 10)

 The silkscreen "10" marks the top-left pad and "1" the bottom-right pad
 in this orientation. Remember JTAG2 elsewhere on the PCB is the wrong header.
```

| JTAG1 | Signal | Status |
| --- | --- | --- |
| **1** | **VTref / VDDS (3.3 V)**
| **2** | **GND**
| **3** | **GND**
| **4** | Unknown
| **5** | **GND**
| **6** | **JTAG_TMSC**
| **7** | **JTAG_TCKC**
| **8** | CC2642 pin 26 (GPIO)
| **9** | CC2642 pin 27 (GPIO)
| **10** | **RESET_N**

##### XDS110 connection

| XDS110 | JTAG1 |
| --- | --- |
| VTref | **1** |
| GND | **2** (or 3/5) |
| TMSC | **6** |
| TCKC | **7** |
| RESET | **10** |

You only need to connect these 5 pins. GND can be any GND from the board.  

##### CC2642R chip pinout

Only needed if the JTAG1 pads are damaged and you have to go to the chip
directly, or to work out where to inject power for step 1j. Pin numbers are
from the TI CC2642R datasheet (SWRS194), 48-pin VQFN **RGZ** package, 7 x 7 mm.

```
                VDDR_RF         VDDR VDDS
                   |              |    |
                   48   47   46   45   44   43   42   41   40   39   38   37
                   |    |    |    |    |    |    |    |    |    |    |    |
                +------------------------------------------------------------+
       RF_P  1 -|                                                            |- 36  DIO_23
       RF_N  2 -|                                                            |- 35  RESET_N   <-- JTAG1 pin 10
    X32K_Q1  3 -|                                                            |- 34  VDDS_DCDC <-- rail
    X32K_Q2  4 -|                                                            |- 33  DCDC_SW   <-- DC/DC inductor
      DIO_0  5 -|  exposed pad underneath                                    |- 32  DIO_22
      DIO_1  6 -|  = GND  (the ONLY ground                                   |- 31  DIO_21
      DIO_2  7 -|   connection to the die)                                   |- 30  DIO_20
      DIO_3  8 -|                                                            |- 29  DIO_19
      DIO_4  9 -|                                                            |- 28  DIO_18
      DIO_5 10 -|                                                            |- 27  DIO_17    --> JTAG1 pin 9
      DIO_6 11 -|                                                            |- 26  DIO_16    --> JTAG1 pin 8
      DIO_7 12 -|                                                            |- 25  JTAG_TCKC <-- JTAG1 pin 7
                +------------------------------------------------------------+
                   |    |    |    |    |    |    |    |    |    |    |    |
                   13   14   15   16   17   18   19   20   21   22   23   24
                   |                                            |    |    |
                 VDDS2                                        VDDS3  |    |
                                                                  DCOUPL  |
                                                                      JTAG_TMSC  <-- JTAG1 pin 6

  TOP VIEW. Pin 1 is the bevelled/dotted corner (top left). Numbering runs
  counter-clockwise: 1-12 left, 13-24 bottom, 25-36 right, 37-48 top.
```

###### Power rails — what you may and may not feed

**Not every pin with VDD in the name is an input.** Step 1j means the
VDDS group only.

| Rail | Pin | Feed it externally? |
| --- | --- | --- |
| **VDDS** | 44 | **Yes** — main supply, 1.8-3.8 V (3.3 V here) |
| **VDDS2** | 13 | **Yes** — DIO supply, same potential as VDDS |
| **VDDS3** | 22 | **Yes** — DIO supply, same potential as VDDS |
| **VDDS_DCDC** | 34 | **Yes** — DC/DC supply, same potential as VDDS |
| VDDR | 45 | **No** — internally generated by the on-chip DC/DC or LDO |
| VDDR_RF | 48 | **No** — internally generated, same as VDDR |
| DCOUPL | 23 | **No** — decoupling output of the internal 1.27 V regulator |
| DCDC_SW | 33 | **No** — switching node, goes to the external inductor |
| **GND** | exposed pad | **Yes** — the pad under the chip is the only ground |

All four VDDS pins must sit at the **same potential**. Driving VDDR, VDDR_RF
or DCOUPL from outside fights the internal regulator and can damage the chip.
Feeding 12Vdc to the charging port as in step 1j lets the board's own regulator
bring these up correctly, which is why that is the easier route.

Note pin numbers here are the **chip** pins, not the JTAG1 header pins — the
two use overlapping numbers and are easy to confuse. JTAG1 pin 8 and 9 land on
chip pins 26 and 27 (DIO_16 / DIO_17), both high-drive capable GPIOs.

3. Prepare to flash the chip

    a. On Windows you need to download the [SmartRF Flash Programmer 2](https://www.ti.com/tool/de-de/download/FLASH-PROGRAMMER-2/1.8.2) to "Force Mass Erase" the CC26xx chip.  
    *For the download you need a TI account.*  
    b. Open the Flash Programmer.  
    c. If you connected everything correctly the XDS110 should show up in the left sidebar.  
    d. Click on the XDS110.  
    e. When you want to connect to the chip the tool will already tell you that the chip is locked and that access to the device is blocked.  
    f. Just click OK.  
    g. It will now warn you that a forced mass erase will be performed.  
    h. "Just" click OK. -> *This will reset only the BLE chip to factory defaults, no other chip or flash is affected.*  
    After a while the bottom progress bar will turn green and show "Success!".  
    On the bottom left the selected target state MUST be "Connected".  
    **If one or both of the previous operations failed you can't continue.**  

    
4. Finally flash the chip

    a. On the page "Main" select "Single" and leave Address untouched if you merged bleboot and bleware into one flashable image with [chwdt/vanmoof-tools](https://github.com/chwdt/vanmoof-tools) OR use "Multiple" if you have just the images. Use address 0x00056000 for bleboot and 0x0 for bleware. If you have a Find My Bike you need 2.4.01, if you do not you can use both, but the factory default is the 1.4.01 version.
    b. Actions "Program" should be preselected. Keep everything else untouched.  
    c. Press the play button.  
    *wait patiently*  
    d. If the flash was successful and you used the modified single image the chip will stay unlocked. If you flashed the original images the chip will become locked again.  
    e. You can check if you can already connect to the bike again. If the flash was successful the chip should already be ready to accept commands. At least it announces the ES3-MAC via BLE.

5. Assembly

    a. Use Loctite on the screws of the 5 big wires — they carry a lot of current and must not work loose. Keep it off the contact faces: the top of the round aluminium post and the round metal plate on the wire. Those must stay clean.  


### Sensors & Components

#### Error 38

Charge module via 12Vdc or charge the module battery directly.  
If you can, check internal resistance to check battery health.  

Otherwise: Replace module battery  

#### Errors 40, 41

Hardware error: button(s) do not work all the time.  

Clean button(s) with contact cleaner or replace button(s) (2€ for 5pcs on AliExpress)  

#### Error 44  

eShifter likely dead. You need to replace all shorted resistors or replace it with a better PCB.  
You can manually shift into the 2nd gear.  
Remove the wheel from the frame, unscrew the eShifter from the wheel. Keep all screws and rings secure!  

In case you want to repair it, I needed  

- 100 ohm size 0603  
- 10 K ohm size 0603  
- 270 K ohm size 0603  
- 51 K ohm size 0603  
- 22 ohm size 0603  

https://www.reddit.com/r/vanmoofbicycle/comments/1k2eapb/vanspoof_a_possible_answer_to_your_44_err_woes/  
https://mikecoats.com/van-spoof-v1-0/  
https://www.reddit.com/r/VanMoofSelfRepair/comments/1f2fcv3/vanmoof_s3_x3_smd_components/  
-> Changing mainware settings with Chris' repo to disable the shifter but keep the display enabled.  

## ⚙️ Motor Repairs (Errors 45-53)

### Error 45

Motor temperature sensor open circuit.  
Inside the motor is a small PCB carrying a temperature probe and a hall sensor.  

Potential impact: **the motor loses its overheat protection.** Normally power is throttled back above 90 °C.
Set condition: motor temperature line reads open circuit, seen while riding.  
Clear condition: sensor reads sane again - but the code only updates **during an actual ride**,
so after a repair take the bike out before assuming it did not work. It will not clear standing still.  

Motor temperature in the app reads ~239 °C instead of blank.  

The hall sensor gives you a free field test: **spin the front wheel and
watch the matrix.** If the speed comes up, the hall line through the boom is still intact. It
feeds the speed display, the odometer and the alarm's movement trigger, and it has **no error
code of its own** — it just dies quietly. Error 45 *and* no speed on the matrix means the boom
is further gone than the temperature sensor alone would tell you.

**If the motor still works, this is the pre-announcement of a failing cable boom!**  
The thermistor wires breaking is stage one, the Hall line and the phase wires are next.  
With bad luck this could ruin the entire bike.  
Only longterm solution is to replace the entire cable boom to the new version.  

## 📶 Connectivity Repairs (Errors 54-58)

### Error 54

SIM Card not detected.  

Potential impact: Err 23 is shown on the Bike, with Ride Pro theft protection no longer active.  
Set condition: Modem responds but no SIM card in slot detected.  
Clear condition: Modem detects valid SIM card. 


Check if the SIM card works in another modem, Phone, Smartphone.    
If the SIM card is dead, which happens, you are out of luck. But you can use another one from another bike.  
The identification is done with the MAC from the BLE chip.  
In MODEM.md are examples on how to check the SIM.  

Errors 56, 57

Error 58

GSM modem might be dead. Check the output from GSMdebug. The good thing is this chip is easy to get.


### Firmware Recovery

#### Mainware

Bike seems dead, BLE connects, no unlock possible.

1. Connect to Debug Port on the Module as shown in README.md
1. Enter bootloader mode  
    a. press ESC until the bootloader is shown  
    b. power down the pcb (remove battery and charger) press ESC while connecting the PCB to power  
3. ea -> erase application
4. ua -> upload application
5. Use YMODEM to upload the firmware 
5. st -> start application

If no bootloader is shown or there is no output at all:  

First check the cabling, module battery, module voltages, resistors. Check if Vcc on the STM32 has voltage!  
Either the STM32 chip is dead or the bootloader got corrupted. Fixing the bootloader is only done via SWD.  
To connect via SWD you need a TAG-Conntect Adapter.   
If the chip is detected with SWD, you are good. If the chip is not even shown in SWD, the chip is likely dead.  

---

## 🚨 Safety Warnings

**⚠️ CRITICAL SAFETY NOTES:**
- Always disconnect power before repairs
- Use proper ESD protection
- Never bypass safety systems, especially the fuse in the BMS

**🔴 STOP IMMEDIATELY:**
- Any burning smell, excessive heat, smoke or flames
- Visible damage to battery as a whole or the battery cells

---

## 🛠️ Required Tools

### Basic Repair
- iFixit screwdriver set
- Multimeter
- Isopropyl alcohol (99%)
- Threadlocker (Loctite)
- Chain lubrication oil
- Gloves

### Advanced Repair
- Oscilloscope
- SPI flash programmer
- Raspberry Pi (for firmware updates and SWD)
- ESD protection equipment
- Sand