## Modem

The Location of the bike is estimated by using the last location and the current cell tower the modem is connected to.  
Using the current time from the cell tower and/or the GNSS mainware uploads these information to ublox1.vanmoof.com.  
AssistNow Live Orbits calculates based on these information the current estimated location and sends over precalculated GNSS Data.  
This saves time and the complex calculation to get the bike location. 


https://www.u-blox.com/en/product/assistnow
https://portal.thingstream.io/pricing/laas/assistnow
https://support.thingstream.io/hc/en-gb/articles/19690127778204-AssistNow-User-guide
https://www.u-blox.com/en/product/u-center
https://www.u-blox.com/en/u-center-2

https://www.mikrocontroller.net/attachment/243876/u-blox-ATCommands_Manual__UBX-13002752_.pdf
https://content.u-blox.com/sites/default/files/products/documents/MultiGNSS-Assistance_UserGuide_%28UBX-13004360%29.pdf
https://content.u-blox.com/sites/default/files/SER-product_Overview_UBX-21029993.pdf

### Some Communication from the Bike to the VanMoof Backend via self-signed Certs.

The Backend only supports
```console
Hexcode  Cipher Suite Name (OpenSSL)       KeyExch.   Encryption  Bits     Cipher Suite Name (IANA/RFC)
 x3c     AES128-SHA256                     RSA        AES         128      TLS_RSA_WITH_AES_128_CBC_SHA256
```

The UUID is without dashes, 32 chars, numbers and chars. No duplication checking.
dist is in kilometers with hectometers. So 5.5 kilometers become 55 here.
Responds with result true.
```console
curl -vk https://bikecomm.vanmoof.com/ping-response \
-H "Content-Type: application/json" \
-d '{"guid":"UUID","statistics":{"batt":95,"mac":"MAC","swv":"1.6.8","dist":37154}}'
```

Message Type(s) unknown
```console
curl -vk https://bikecomm.vanmoof.com/bike-message 
-H "Content-Type: application/json" 
-d '{"mac_address":"MAC","message_type":"","message_data":""}'
```

/upload expects UBlox Data. => InvalidUBloxDataException

### m2m.vanmoof.com (SMS the bike sends)

```console
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

GSM: sms= %d
%02X:%02X:%02X:%02X:%02X:%02X 'mac_address':'%s','message_type':'%s','message_data':'%s'
'guid':'%s','statistics':{'batt':%d,'mac':'%s','swv':'%s','dist':%d}


### Modem
To Test!
```console
gsmstart
Start GSM
gsmdebug
Connect to UART2
Modem powering on..
ATI
SARA-G350-02S-01

OK
ATI9
08.90,A01.13

OK
AT+CGSN
357796113371337

OK
AT+GMI
u-blox

OK
AT+GMM
SARA-G350

OK
AT+GMR
08.90

OK
AT+CPIN?
+CPIN: READY

OK
AT+CCID
+CCID: 89314404000813371337

OK
AT+CIMI
204047113371337

OK
AT+COPS?
+COPS: 0,0,"Vodafone.de"

OK
AT+CREG?
+CREG: 0,5

OK
AT+UCELLINFO
ERROR
AT+CCLK?
+CCLK: "04/01/01,00:11:33+00"

OK
AT+CTZU?
+CTZU: 0

OK
AT+CTZR?
+CTZR: 0

OK
AT+CREG?
+CREG: 0,5

OK
AT&V
ACTIVE PROFILE:
&C1, &D1, &S1, &K3, E1, Q0, V1, X4, S0:000, S2:043, S3:013, S4:010, S5:008,
S7:060, +CBST:7,0,1, +CRLP:61,61,48,7, +CR:0, +CRC:0, +IPR:0,
+COPS:0,0,FFFFF, +ICF:0,0, +UPSV:0, +CMGF:0, +CNMI:1,0,0,0,0, +USTS:0, +UTPB:0, +CSNS:0
AT+CSCA?
+CSCA: "+316540967011",145

OK

```

It might be possible to update the Modem Firmware via ```AT+ULOAD```.

From mainware

```console
+CREG: Network registration status
Error
+CCID: ICCID
+CPIN: 
+CMGF: 0
+UPSND:
+UUHTTPCR:
+CSQ:
+CPMS:
+CMGL:
+UULOC:
+CMGR:
+UGSRV:
+UGAOP:
+UPSD:
AT+UGAOP?
AT+UGAOP="%s",%s,1000,0
AT+UGSRV
AT+UGSRV="%s","%s","%s",14,4,1,65,0,15
AT+ULOC=2,2,1,120,1000
AT+UHTTP=0,1,"%s"
AT+UHTTP=0,5,%s
AT+UHTTP=0,6,1
AT+UHTTPC=0,5,"/upload","https","%s",1
AT+UHTTPC=0,5,"/bike-message","https","{%s}",6,"application/json"
AT+UHTTPC=0,5,"/ping-response","https","{%s}",6,"application/json"
AT+UPSDA=0,4
AT+UPSD=0,1
AT+UPSD=0,1,"%s"
AT+UPSND=0,8
AT+UPSDA=0,3
AT+CMGS="%s"
AT+CMGL
AT+CMGR=%d
AT+CMGD=%d
AT+CMGF?
AT+CMGF=1
AT+CPMS?
AT+CPWROFF
AT+CGMI
AT+CGMM
AT+CGMR
AT+CGSN
AT+CPIN?
AT+CPIN="%s"
AT+CIMI
AT+CCID
AT+CSQ
AT+CREG?
AT+COPS?
AT+CEREG?
```

imei=%s&rmc=$

### Sending an SMS from the Bike?
```console
AT+CMEE=2
OK
AT+CMGF=1
OK
AT+CMGS="+49152568991337"<CR>
> Hello from my VanMoof!
<CTRL-Z>
+CMGS: 4

OK
```

All Modem Commands
```console
AT+CLAC
ATD

ATDL

ATD>

AT&A

AT&B

AT&C

AT&D

AT&E

AT&F

AT&H

AT&I

AT&K

AT&M

AT&R

AT&S

AT&V

AT&W

AT&Y

ATA

ATB

ATE

ATH

ATI

ATL

ATM

ATO

ATP

ATQ

ATS0

ATS10

ATS12

ATS2

ATS3

ATS4

ATS5

ATS6

ATS7

ATS8

ATT

ATV

ATX

ATZ

AT\Q

AT+CACM

AT+CALA

AT+CALD

AT+CALM

AT+CAMM

AT+CAOC

AT+CBST

AT+CCFC

AT+CCHC

AT+CCHO

AT+CCID

AT+CCLK

AT+CCUG

AT+CCWA

AT+CCWE

AT+CECALL

AT+CEER

AT+CFUN

AT+CGACT

AT+CGATT

AT+CGCLASS

AT+CGDATA

AT+CGDCONT

AT+CGED

AT+CGEREP

AT+CGLA

AT+CGMI

AT+CGMM

AT+CGMR

AT+CGPADDR

AT+CGQMIN

AT+CGQREQ

AT+CGREG

AT+CGSMS

AT+CGSN

AT+CHLD

AT+CHUP

AT+CIMI

AT+CIND

AT+CLAC

AT+CLCC

AT+CLCK

AT+CLIP

AT+CLIR

AT+CLVL

AT+CMEE

AT+CMER

AT+CMGD

AT+CMGF

AT+CMGL

AT+CMGR

AT+CMGS

AT+CMGW

AT+CMOD

AT+CMSS

AT+CMUT

AT+CMUX

AT+CNAP

AT+CNMA

AT+CNMI

AT+CNUM

AT+COLP

AT+COLR

AT+COPN

AT+COPS

AT+CPAS

AT+CPBF

AT+CPBR

AT+CPBS

AT+CPBW

AT+CPIN

AT+CPMS

AT+CPOL

AT+CPUC

AT+CPWD

AT+CPWROFF

AT+CR

AT+CRC

AT+CREG

AT+CRES

AT+CRLA

AT+CRLP

AT+CRSL

AT+CRSM

AT+CSAS

AT+CSCA

AT+CSCB

AT+CSCS

AT+CSDH

AT+CSGT

AT+CSIM

AT+CSMP

AT+CSMS

AT+CSNS

AT+CSQ

AT+CSSN

AT+CSTA

AT+CTFR

AT+CTZR

AT+CTZU

AT+CUAD

AT+CUSD

AT+CUUS1

AT+FAA

AT+FAP

AT+FBO

AT+FBS

AT+FBU

AT+FCC

AT+FCLASS

AT+FCQ

AT+FCR

AT+FCS

AT+FCT

AT+FDR

AT+FDT

AT+FEA

AT+FFC

AT+FFD

AT+FHS

AT+FIE

AT+FIP

AT+FIS

AT+FIT

AT+FKS

AT+FLI

AT+FLO

AT+FLP

AT+FMI

AT+FMM

AT+FMR

AT+FMS

AT+FND

AT+FNR

AT+FNS

AT+FPA

AT+FPI

AT+FPP

AT+FPS

AT+FPW

AT+FRQ

AT+FRY

AT+FSA

AT+FSP

AT+GCAP

AT+GMI

AT+GMM

AT+GMR

AT+GSN

AT+ICF

AT+IFC

AT+IPR

AT+STKENV

AT+STKPRO

AT+STKPROF

AT+STKTR

AT+VTD

AT+VTS

AT+UAEC

AT+UAGC

AT+UANTR

AT+UAUTHREQ

AT+UBANDSEL

AT+UBIP

AT+UCALLSTAT

AT+UCD

AT+UCELLLOCK

AT+UCGOPS

AT+UCLASS

AT+UCMGL

AT+UCMGR

AT+UCMGS

AT+UCMGW

AT+UCSD

AT+UCSDA

AT+UCSND

AT+UCTS

AT+UDBF

AT+UDCONF

AT+UDELFILE

AT+UDNSRN

AT+UDOPN

AT+UDTMFD

AT+UDWNFILE

AT+UDYNDNS

AT+UECALLDATA

AT+UECALLSTAT

AT+UECALLTYPE

AT+UECALLVOICE

AT+UEONS

AT+UFACTORY

AT+UFTP

AT+UFTPC

AT+UFTPER

AT+UFWUPD

AT+UGAOF

AT+UGAOP

AT+UGAOS

AT+UGCNTRD

AT+UGCNTSET

AT+UGGGA

AT+UGGLL

AT+UGGSA

AT+UGGSV

AT+UGIND

AT+UGPIOC

AT+UGPIOR

AT+UGPIOW

AT+UGPRF

AT+UGPS

AT+UGRMC

AT+UGSRV

AT+UGTMR

AT+UGUBX

AT+UGVTG

AT+UGZDA

AT+UHFP

AT+UHTTP

AT+UHTTPAC

AT+UHTTPC

AT+UHTTPER

AT+UI2S

AT+UIPCHGN

AT+ULOC

AT+ULOCAID

AT+ULOCCELL

AT+ULOCGNSS

AT+ULOCIND

AT+ULSTFILE

AT+UMGC

AT+UMSM

AT+UNFM

AT+UNFMCONF

AT+UPAR

AT+UPINCNT

AT+UPING

AT+UPLAYFILE

AT+UPSD

AT+UPSDA

AT+UPSND

AT+UPSV

AT+URDBLOCK

AT+URDFILE

AT+URNG

AT+URPM

AT+USAR

AT+USECMNG

AT+USECPRF

AT+USER

AT+USGC

AT+USIMLCK

AT+USIO

AT+USMTP

AT+USMTPC

AT+USMTPER

AT+USMTPM

AT+USOAO

AT+USOCL

AT+USOCO

AT+USOCR

AT+USOCTL

AT+USODL

AT+USOER

AT+USOGO

AT+USOLI

AT+USORD

AT+USORF

AT+USOSEC

AT+USOSO

AT+USOST

AT+USOWR

AT+USPM

AT+USTN

AT+USTOPFILE

AT+USTS

AT+UTEST

AT+UTGN

AT+UTPB

AT+UUBF

OK
```