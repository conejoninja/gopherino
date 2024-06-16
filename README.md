# Gopherino

Gopherino is a robot based on the [Maqueen v4.1 platform](https://wiki.dfrobot.com/micro_Maqueen_for_micro_bit_SKU_ROB0148-EN#target_6) and powered by a [BBC Microbit](https://wiki.dfrobot.com/micro_Maqueen_for_micro_bit_SKU_ROB0148-EN#target_6). It has several sensors onboard, RGB LEDs and two motors to move. As well as several expasion connectors.

![Gopherino](gopherino.jpeg)

## Sensors and pins
- Infrared grayscale sensors (for line tracking) at P13 & P14
- 4x RGB LEDS (WSS2812) at P15
- 2x RED LEDs at P8 & P12
- Infrared sensor at P16
- Buzzer at P0
- HC-SR04 ultrasonic sensor (connected) at P1 & P2
- Motors runs on I2C


## Ideas
The current code in this repository for Gopherino is quite simple, as it only uses the HC-SR04 ultrasonic sensor and the motors to avoid obstacles. So the possibilities to expand it are limitless. Here are some ideas.

- Use the infrared grayscale sensors to make it  a robot following lines
- Make some noise with the buzzer (perhaps when it encounters and obstacle)
- Set the room ambient with its RGB LEDs
- Accept commands via the IR sensor (you might need a IR emitter, like a tv remote or a flipper zero)
- Take pictures with it, it's super cute!





## License

Public domain. Feel free to use however you wish. Attribution is appreciated though.
