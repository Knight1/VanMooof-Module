# VanMooof ES3 Firmware

## Firmware Changelogs
#### 1.1.15
- Initial version

#### 1.1.18
- Alarm tweaks
- Improved riding behavior for Japan

#### 1.2.0
- Improves the firmware update process
- Fix for persistent battery charging issues
- Fix for potential battery error 17, 19, and 20
- Several other small bug fixes

#### 1.2.2
- Fixes discoverability issue for affected Android phones

#### 1.2.4
- Improved light sensor settings
- A tuned amplifier for better speaker performance
- A more efficient software update process
- Several other small bug fixes

#### 1.2.7
- Improved light sensor settings

#### 1.4.5
- Resetting restores factory setting
- Optimized alarm timings
- Fine-tuned tracking mode
- Improved E-shifter accuracy
- Battery levels are more reliable
- Multiple bug fixes

#### 1.4.7 (Beta Firmware)
motor:.20,shifter:0.237
region 
#### 1.6.8 16. April 2021 
ble:X.4.1
- Prepared for PowerBank
- Failed updates automatically restart at shut down
- Resetting bike restores factory settings
- Fine-tuned tracking mode
- Improved E-shifter accuracy
- Various small bug fixes

#### 1.6.13 15. Juni 2021
- Prepared for PowerBank
- Alarm sensitivity optimization
- Improved battery management
- Various small bug fixes

#### 1.7.1 2. August 2021
- manual shifting using the handlebar buttons
- change motor assistance level while riding
- alarm stays active when the lock is unintentionally unlocked
- prevents rare accidental battery drains when stationary
- fix for possible update loop
- fix for error 13 and 16
- various small bug fixes

#### 1.7.2 27. August 2021
- manual shifting using the handlebar buttons
- change motor assistance level while riding
- alarm stays active when the lock is unintentionally unlocked
- prevents rare accidental battery drains when stationary
- fix for possible update loop
- fix for error 12 and 13
- various small bug fixes

#### 1.7.3 12. Oktober 2021
1.7.2 + improvements
#### 1.7.6 1. November 2021
ble:X.4.1,motor:.22,shifter:0.237,bms:1.20.1
- improved interaction for Back Up unlock code
- optimised power management at lower battery levels
- fix for possible boost issues after power level change
- fix for Error 16
- various small bug fixes.
#### 1.8.1 22. Februar 2022
ble:X.4.1,motor:.22,shifter:0.237,bms:1.20.1
- Removal of testing module settings
- Gear settings not reset after a firmware update
- Deep sleep after 14 days of inactivity
- Allows motor support in case of error 57

#### 1.8.2 ✅ 16 June 2022
#### 1.9.1 11. Juni 2023
(Part 1)
- You can now unlock your bike with the Boost button, as well as the Bell button.
- When your bike is standing still, it will automatically shut down after 7 minutes. This is longer than before.
- A new update to the battery means that whenever your bike hasn’t been used for a while, you’ll need to press the Power button to wake up the bike.
#### 1.9.3 ✅ 12 June 2023
(Part 2)
- Other important updates in battery management, firmware, error management and sounds.

## Known Firmware File SHA512 Sums

### Mainware (muco) boot
**unknown**
cf8f1e480ed729360a4a83643fb41f6f4e6d085f0ad5faca24eacb7afc0339a6bdcd0657d6a42b9f624e822bea6d86cb3db10faeda6c6e2e0990182c8a309575

### Mainware
**1.08.02**
66cee63020ea35447fc7dcf41b61300715937f0f19d02ffdc1626ca0e8356fe00fff57fab0ef043077829129c4b66c40bd823a9e6ae0325c4d048227a5664587
```
version 010802f4
CRC 0x4e0f9854
length 0x00035bf0
date May  9 2022
time 10:58:01
```
**1.09.01** 
c891be8c0a81a5901143343ba8b65bec4354d81886c63df75cb82410dbaf9c6261e159c110328058fdd0d2492ac21bbdaf6b45d64b2e128c231f9f533c545502
```
version 010901f4
CRC 0x1016bdd5
length 0x0002fcac
date May  9 2023
time 14:11:15
```

**1.09.03** 
52780d5fb984d954cc81a4ab2f72e612639b6573ea1c250bc96c0ee0707444a6fef7f82c6c3a347813ab14ca661116aef51a45eb9852ddf9cd53e83f16b35256
```
version 010903f4
CRC 0x76c1ab9d
length 0x0002fcc8
date Jun  7 2023
time 07:21:48
```

### Motorware
**S.0.00.15**
d36394fd1ff33a18baf4efac6fd754189c392aa217a6dc155f3316af213b65c16da76b345f64e1e613d04719496c6224846c5aad18d39822312a0db34e8b7e1f
```
Missing
```

**S.0.00.22** 
ce5815d55366a10decf724224d03a44d2d43ce3093de4d9f85a5ea646594b3cec59dd3a82227146f8069f41fcf381e341f391ccbf1f339c296cd768368a177cc
```
version 000016a1
CRC 0x9e5dd658
length 0x0000f118
date  03 2021
time 00:48:35
```

### Shifterboot
**unknown**
b08403daf0ec4fec7e80a4e926a3a2e953471bb41aab7e939242ce81b14f556390edad5b2955ea0bbf211dbf5175e2e6eaf59935c200e94801259489167c88bc

### Shifterware
**0.237** 
8f454dfc1e600dfeae772465dd9791cde1b7588be22f7d88e88e61c9708634173a730be85ba19214d6c4544576ebeb8ea7e51e2686a163b77f7693292da97409
```
version 00ed02c1
CRC 0x1e8eb125
length 0x00002ea8
date Oct 23 2020
time 14:09:11
```

### BLEBoot
**1.0.0**
efcef2f649aca663b6343daf1efda7852bdc59497c1bf3551a0061648ea72686111e4d74bedc92d0d20470131a8d5e83f1a666d2b7daa9f410c75f184dc5d601

### BLEware
**1.04.01** 
118084995f7423cf8b1c5589d49b20f203c06a4116213b4264a4c30d25060fee2fe057e1906e8a3ec9ab5323a02b2f72ee454fbc2c9cc7e6ca550ab71abcfe52

**2.04.01** 
467f425f8ff329204876159697a71e04dec2b9fc7336892d233f68d7ce8ab8a4eb9b3dea506d5f885008a602301eb9a2ecbba66327379eb860115edd37a3057c

### BMSBoot
**unknown** 
cd2fdb29adc315da8b99d81d0ac18cacf13fbe0399a3763bc737df8b214fd6628804c1b55929da3d8a0f906ae8fc00884e108755152f6a840acfcb17460b3bcf

### Batteryware
**1.14.01** 
dc3c1e3e731936f3c20dd6432e9a2b2855e7a699a179488e3d878438a2663d779d8640df2692e2008a885a0917fb270a95738ce75f731503ec70e7f3a6c72e02
```
version 011401b1
CRC 0x27f3de41
length 0x000147e4
date Nov  1 2021
time 14:32:13
```

**1.17.01** 
e74f4db508323c486bcd149b972a44d8116a8fa1b9daa50b760a2ba70ac87745677cc3846dabdc471950c0470e915d749f42c4218554eb1171544e93bc6301de
```
version 011701b1
CRC 0x2e0150da
length 0x00015610
date 
time
```