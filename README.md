# VanMooof-Module 

### How to get started?

You need the backside of the PCB from the Module to dump the SPI Flash which contains

- ble key
- fmn key
- all Firmwares
  - Mainware
  - Batteryware
  - Shifterware
  - bleware
- Logs

1. unlock bike and remove Module from the Frame
   2. if you do not unlock the bike, the Alarm stays on and will annoy you. I used duct tape to cover the speaker.
2. open Module and unscrew all internal screws of the PCB to remove the PCB.
3. On the backside of the PCB is a Winbond MX25L51245G with 512Mbit, so 64MB of Flash
4. Dump that Flash with an 16 Pin! SPI Flash Chip clamp and a Pi
   5. I used an Raspberry Pi Zero v1.1. There you have to enable the SPI Interface with raspi-config 
```console
# sudo flashrom -p linux_spi:dev=/dev/spidev0.0 -r rom.rom
flashrom v1.2 on Linux 6.1.21+ (armv6l)
flashrom is free software, get the source code at https://flashrom.org

Using clock_gettime for delay loops (clk_id: 1, resolution: 1ns).
Using default 2000kHz clock. Use 'spispeed' parameter to override.
Found Macronix flash chip "MX66L51235F/MX25L51245G" (65536 kB, SPI) on linux_spi.
Reading flash... done.
```