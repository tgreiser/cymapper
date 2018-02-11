# cymapper
Map LED positions with a webcam. Created in association with CymaSpace - http://www.cymaspace.org/

Requires: gocv, opencv, teensy/arduino driver, webcam

Sequentially activate addressable LEDs which are attached to a teensy (via USB/serial). Detect bright spots with the webcam to map positional locations of each LED pixel. Output a CSV of LED addresses and relative coordinates that can be used for pixel mapping.

### Arguments

Still being implemented:

- leds - per strip
- strips - number of pins connected
- radius - radius of the gaussian blur used for noise reduction
- width - of your video stream
- height - of your video stream
- border - unused space to leave around the outer pixels

### LED setup

ws8211/8212 connected to a teensy (or compatible). Load the teensy with Lucas Morgan's ledPixelController_460 (see: https://www.derivative.ca/forum/viewtopic.php?f=4&t=6654&start=30#p28824)
