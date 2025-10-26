# VanMoof S3/X3 Error Codes - Complete Reference

*A simple guide to understand error codes*

## 📋 Error Code Categories Summary

- **🔋 Battery (0-21):** Battery
- **📡 System (22-44):** Communication, sensors, buttons
- **⚙️ Motor (45-53):** Motor
- **📶 Connectivity (54-58):** GSM Modem
- **🏭 Special (60+):** Factory and diagnostic modes

## 🔋 Battery & Power Issues (Errors 1-21)
All errors are Battery related

### Critical Battery Errors

| Code   | What It Is    | What It Means                             |
|--------|---------------|-------------------------------------------|
| **0**  | **BAT_SCP**   | Short circuit protection                  |
| **2**  | **BAT_PDSCP** | Pre-discharge short circuit protection    |
| **3**  | **BAT_PDOCP** | Pre-discharge overload circuit protection |
| **16** | **BAT_FP**    | Permanent fault (fuse triggered)          |

### Voltage Problems

| Code  | What It Is   | What It Means                       |
|-------|--------------|-------------------------------------|
| **4** | **BAT_UVP2** | Under-voltage protection (Level 2)  |
| **5** | **BAT_UVP1** | Under-voltage protection (Level 1)  |
| **6** | **BAT_OVP2** | Over-voltage protection (Level 2)   |
| **7** | **BAT_OVP1** | Over-voltage protection (Level 1)   |

### Charging & Current Issues

| Code   | What It Is            | What It Means                               |
|--------|-----------------------|---------------------------------------------|
| **8**  | **BAT_COCP1**         | Charge over-current protection (Level 1)    |
| **9**  | **BAT_COCP2**         | Charge over-current protection (Level 2)    |
| **10** | **BAT_DOCP2**         | Discharge over-current protection (Level2   |
| **11** | **BAT_DOCP1**         | Discharge over-current protection (Level 1) |
| **21** | **NO_CHARGE_CURRENT** | Charger connected but no current detected   |

### Temperature Problems

| Code   | What It Is   | What It Means                          |
|--------|--------------|----------------------------------------|
| **1**  | **BAT_MOTP** | MOSFET over-temperature protection     |
| **12** | **BAT_CUTP** | Charge Under Temperature protection    |
| **13** | **BAT_COTP** | Charge Over Temperature protection     |
| **14** | **BAT_DUTP** | Discharge Under Temperature protection |
| **15** | **BAT_DOTP** | Discharge Over Temperature protection  |

### System Communication

| Code   | What It Is                | What It Means                          |
|--------|---------------------------|----------------------------------------|
| **17** | **BATTERY_NO_DSG**        | Can't activate discharge mode          |
| **18** | **BATTERY_MISSING**       | Battery not detected (KEY_IN not High) |
| **19** | **BATTERY_COMMUNICATION** | Can't talk to battery                  |
| **20** | **BATTERY_NO_OUTPUT**     | Battery management shut down           |

## 📡 Communication & System Issues (Errors 22-44)

### Core Communication

| Code   | What It Is              | What It Means                  |
|--------|-------------------------|--------------------------------|
| **22** | **MOTOR_COMMUNICATION** | Can't talk with TI motor chip  |
| **23** | **BLE_COMMUNICATION**   | Can't talk with Bluetooth chip |

### OAD Errors (Over-Air Download)

| Code   | What It Is       | What It Means                |
|--------|------------------|------------------------------|
| **24** | **OAD_ABORT**    | Transfer aborted             |
| **25** | **OAD_CRC**      | crc32 error in uploaded file |
| **26** | **OAD_TRANSFER** | Transfer timeout             |
| **27** | **OAD_PACK**     | Pack is missing or invalid   |

### ICF Errors (Internal Component Flash)

| Code   | What It Is       | What It Means                   |
|--------|------------------|---------------------------------|
| **28** | **ICF_TIMEOUT**  | Timeout during flash            |
| **29** | **ICF_HEADER**   | Header not found                |
| **30** | **ICF_NO_FILES** | No Files in Pack                |
| **31** | **ICF_ERASE**    | Flash Erase failed              |
| **32** | **ICF_WRITE**    | Flash Write failed              |
| **33** | **ICF_CRC**      | CRC32 invalid from File in Pack |

### PGM Errors (Controller Programming)

| Code   | What It Is          | What It Means                      |
|--------|---------------------|------------------------------------|
| **34** | **PGM_MOTORWARE**   | Motor Update failed                |
| **35** | **PGM_BATTERYWARE** | Battery Update failed              |
| **36** | **PGM_SHIFTERWARE** | Shifter Update failed              |
| **37** | **PGM_BLEWARE**     | Bluetooth Controller Update failed |

### Sensors & Components

| Code   | What It Is                | What It Means                       |
|--------|---------------------------|-------------------------------------|
| **38** | **INTERNAL_BATTERY**      | STVC3115 read error                 |
| **39** | **READ_LIGHT_SENSOR**     | CM3232 I2C Light sensor read error  |
| **40** | **STUCK_HORN_BUTTON**     | Horn button stuck                   |
| **41** | **STUCK_BOOST_BUTTON**    | Boost button stuck                  |
| **42** | **KL_COIL_MISSING**       | Kick-lock coil not detected         |
| **43** | **SHIFTER_NOT_IN_GEAR**   | Gear shifting failed after 10 tries |
| **44** | **SHIFTER_COMMUNICATION** | Can't talk to shifter               |

## ⚙️ Motor Issues (Errors 45-53)

| Code   | What It Is                   | What It Means                                    |
|--------|------------------------------|--------------------------------------------------|
| **45** | **MOTOR_MOTOR_CABLE**        | Motor cable disconnected                         |
| **46** | **MOTOR_MOTOR_OVER_CURRENT** | Motor Driver DRV8301 over-current                |
| **48** | **MOTOR_CONTROLLER_ERROR**   | Motor controller error                           |
| **49** | **MOTOR_CURRENT_ERR**        | Current offset calculation deviated from default |
| **50** | **MOTOR_VOLTAGE_ERR**        | Voltage offset calculation incorrect             |
| **51** | **MOTOR_DERATING**           | Power limited due to high temperature            |
| **52** | **MOTOR_TORQUE_SENSOR_FAIL** | Torque sensor failed                             |
| **53** | **MOTOR_NOT_READY**          | Motor not ready                                  |

## 📶 Connectivity Issues (Errors 54-58)

| Code   | What It Is        | What It Means                                   |
|--------|-------------------|-------------------------------------------------|
| **54** | **NO_SIMCARD**    | SIM card not detected                           |
| **55** | **I2C3_FAIL**     | Communication bus error on startup              |
| **56** | **CCID_SIMCARD**  | Wrong SIM card ICCID                            |
| **57** | **READ_SIMCARD**  | SIM detected but can't communicate with VanMoof |
| **58** | **GSM_MODEM**     | Modem startup failed                            |

## 🏭 Special Modes (Error 60+)

| Code   | What It Is        | What It Means            |
|--------|-------------------|--------------------------|
| **60** | **FACTORY_MODE**  | factory firmware running |

## 🔧 Need Repair Instructions?

**For detailed repair procedures and troubleshooting steps, see [REPAIRS.md](REPAIRS.md)**

*Most errors can be resolved with basic troubleshooting. When in doubt, consult a VanMoof certified technician.*

