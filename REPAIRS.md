# VanMoof S3/X3 Error Repairs & Fixes

*Detailed repair procedures and troubleshooting steps for VanMoof bike error codes*

## 🔋 Battery & Power Repairs (Errors 0-21)

The first Thing I recommend is setting the Lights to always on if the Module still has power.
When no Light is outputted, the BMS is not outputting any Voltage.

### Critical Battery Errors

Main Errors 0, 6, 9, 10, 16
Most likely also displayed 17, 20

Fuse SFK40-45 likely open line. 
Check fuse. Check Pack and Cell Voltages. Check for visual damage & burns. Reset BMS.
More in the dedicated BMS Repo. Ask me. Not for everyone.

### Voltage Problems (Battery and bike dead)

Errors 4, 5
When the Module is dead, you will not see them.

Battery is likely sitting at 4.2V Pack Voltage. You need to charge the Battery manually. 
Also, the Module battery is likely empty.

### Temperature Problems

Errors 12 - 15.

Check Temperature Sensors in the Battery if the problem persists in good temperature conditions.

Check Manual for resistance Values and pls add them.

### System Communication

Error 19

There is a Chip on the Module which translates the BMS and the eShifter Communication. 
When this Chip goes bad it might show this error. 

Another Problem can be that the STM32 on the BMS does enable output but is not communicating or can not read cell voltages.
You need to check the capacitors going from the cell monitor lines.

Another Problem can be that the STM32 on the BMS bootloops. In this Case you will not get any power.

### Charging & Current Issues

Error 21

Make sure the Charger is outputting 37 - 43 Volts and the battery is not at 80% if Charge limit is set.

## 📡 Communication & System Repairs (Errors 22-44)

### Communication

Error 23

TI (BLE) Chip Firmware corrupted, Chip dead, main resistor on PCB open line.

### Sensors & Components

Error 38

Charge Module via 12Vdc or charge the module battery directly.
If you can, check internal resistance to check battery health.

Otherwise: Replace Module Battery

Errors 40, 41
Hardware Errors: Button(s) do(es) not work all the Time.

Clean Button(s) with contact cleaner or replace Button(s) (2€ for 5pcs on AliExpress)

Error 44

eShifter likely dead. You need to replace all shorted Resistors or replace it with a better PCB.
You can manually shift the Gear into the 2th Gear.
Remove the Wheel from the frame, unscrew the eShifter from the Wheel. Keep all the Screws and Rings secure!


## ⚙️ Motor Repairs (Errors 45-53)

Error 45

If the motor works this is the pre-announcement of a failing cable boom.
Good luck!

## 📶 Connectivity Repairs (Errors 54-58)

Errors 54, 56, 57

SIM Card Error.


In the modem.md are examples on how to check the SIM. 
Check if the SIM Card works in another Modem. The PIN is in the Firmware. 
If the SIM Card is dead which happens you are out of luck. But you can use another one from another Bike.
The Identification is done with the mac from the BLE Chip.

Errors 58. 

GSM Modem might be dead. Check the Output from GSMdebug. The Good thing is this Chip is easy to get.


### Firmware Recovery

1. Enter bootloader mode (press ESC until the Bootloader is shown)
2. Use Y-modem to upload firmware

If no bootloader is shown or there is no output at all. 
First check the cabling, module battery, Module Voltages, resistors. Check if Vcc on the STM32 has Voltage!
Either the STM32 Chip is dead or the Bootloader got corrupted. Fixing the Bootloader is only done via SWD. 
If the Chip is detected with SWD you are good. If the Chip is not even show in 

---

## 🚨 Safety Warnings

**⚠️ CRITICAL SAFETY NOTES:**
- Always disconnect power before repairs
- Use proper ESD protection
- Never bypass safety systems especially the Fuse in the BMS

**🔴 STOP IMMEDIATELY:**
- Any burning smell, excessive heat. smoke or flames
- Visible damage to battery at whole or the battery cells.

---

## 🛠️ Required Tools

### Basic Repair
- iFixit screwdriver set
- Multimeter
- Isopropyl alcohol (99%)
- Threadlocker (Loctite)
- Chain Lubrication Oil
- Gloves

### Advanced Repair
- Oscilloscope
- SPI flash programmer
- Raspberry Pi (for firmware updates and SWD)
- ESD protection equipment
- Sand