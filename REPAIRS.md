# VanMoof S3/X3 Error Repairs & Fixes

*Detailed repair procedures and troubleshooting steps for VanMoof bike error codes*

## 🔋 Battery & Power Repairs (Errors 0-21)

The first thing I recommend is setting the lights to always on if the module still has power.
When no light is outputted, the BMS is not outputting any voltage.

### Critical Battery Errors

Main Errors 0, 6, 9, 10, 16
Most likely also displayed 17, 20

Fuse SFK40-45 likely open line. 
Check fuse. Check pack and cell voltages. Check for visual damage & burns. Reset BMS.
More in the dedicated BMS Repo. Ask me. Not for everyone.

### Voltage Problems (Battery and bike dead)

Errors 4, 5
When the Module is dead, you will not see them.

Battery is likely sitting at 4.2V pack voltage. You need to charge the battery manually. 
Also, the module battery is likely empty.

### Temperature Problems

Errors 12 - 15.

Check temperature sensors in the battery if the problem persists in good temperature conditions.

Check manual for resistance values and please add them.

### System Communication

Error 19

There is a chip on the module which translates the BMS and the eShifter communication. 
When this chip goes bad it might show this error. 

Another problem can be that the STM32 on the BMS does enable output but is not communicating or cannot read cell voltages.
You need to check the capacitors going from the cell monitor lines.

Another problem can be that the STM32 on the BMS bootloops. In this case you will not get any power.

### Charging & Current Issues

Error 21

Make sure the charger is outputting 37 - 43 volts and the battery is not at 80% if charge limit is set.

## 📡 Communication & System Repairs (Errors 22-44)

### Communication

Error 23

TI (BLE) chip firmware corrupted, chip dead, main resistor on PCB open line.

### Sensors & Components

Error 38

Charge module via 12Vdc or charge the module battery directly.
If you can, check internal resistance to check battery health.

Otherwise: Replace module battery

Errors 40, 41
Hardware errors: Button(s) do(es) not work all the time.

Clean button(s) with contact cleaner or replace button(s) (2€ for 5pcs on AliExpress)

Error 44

eShifter likely dead. You need to replace all shorted resistors or replace it with a better PCB.
You can manually shift the gear into the 2nd gear.
Remove the wheel from the frame, unscrew the eShifter from the wheel. Keep all screws and rings secure!


## ⚙️ Motor Repairs (Errors 45-53)

Error 45

If the motor works this is the pre-announcement of a failing cable boom.
Good luck!

## 📶 Connectivity Repairs (Errors 54-58)

Errors 54, 56, 57

SIM card error.


In the modem.md are examples on how to check the SIM. 
Check if the SIM card works in another modem. The PIN is in the firmware. 
If the SIM card is dead which happens you are out of luck. But you can use another one from another bike.
The identification is done with the MAC from the BLE chip.

Errors 58. 

GSM modem might be dead. Check the output from GSMdebug. The good thing is this chip is easy to get.


### Firmware Recovery

1. Enter bootloader mode (press ESC until the Bootloader is shown)
2. Use Y-modem to upload firmware

If no bootloader is shown or there is no output at all: 
First check the cabling, module battery, module voltages, resistors. Check if Vcc on the STM32 has voltage!
Either the STM32 chip is dead or the bootloader got corrupted. Fixing the bootloader is only done via SWD. 
If the chip is detected with SWD you are good. If the chip is not even shown in SWD, the chip is likely dead. 

---

## 🚨 Safety Warnings

**⚠️ CRITICAL SAFETY NOTES:**
- Always disconnect power before repairs
- Use proper ESD protection
- Never bypass safety systems especially the fuse in the BMS

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