# cymapper
Map LED positions with a webcam. Created in association with CymaSpace - http://www.cymaspace.org/

Requires: gocv, opencv, teensy/arduino driver, webcam

Sequentially activate addressable LEDs which are attached to a teensy (via USB/serial). Detect bright spots with the webcam to map positional locations of each LED pixel. Output a CSV of LED addresses and relative coordinates that can be used for pixel mapping.

### LED setup

ws8211/8212 connected to a teensy (or compatible). Make note of the COM port and load the teensy with Lucas Morgan's ledPixelController_460 (see: https://gist.github.com/tgreiser/243b9d6152b0196bdea8e8465b83a00e or https://www.derivative.ca/forum/viewtopic.php?f=4&t=6654&start=30#p28824)

### Camera Test

Run to test and position your webcam.

```
  -device-id int
        Device ID of your webcam

> env.cmd
> go run cmd\camtest\main.go -device-id=0
```


### Mapping

cmd/cameramap
```
  -com string
        COM port for teensy (default "COM8")
  -delay-ms int
        Number of milliseconds to pause on each LED (default 1000)
  -device-id int
        Device ID of your webcam
  -file string
        Filename for the tsv output (default "output.tsv")
  -leds int
        Number of LEDs per strip (1-10000) (default 460)
  -pins int
        Number of pins which have LEDs connected (default 8)
  -radius int
        Radius of the gaussian blur used for noise reduction (default 21)
  ```
  
```
> env.cmd
> go run cmd\cameramap\main.go -pins=1 -leds=50 -com=COM9
# saves to output.tsv
```

### Resize

```
  -border int
        Unused space to leave around the outer pixels (default 4)
  -file string
        Filename for the tsv output (default "remapped.tsv")
  -flip-x
        Flip the image along the X axis
  -flip-y
        Flip the image along the Y axis (default true)
  -vheight int
        Height of the video stream you will be mapping (default 720)
  -vwidth int
        Width of the video stream you will be mapping (default 1280)

# on windows
> env.cmd
> type output.tsv | go run cmd\resize\main.go -vwidth 640 -vheight 480

# on linux/osx
> env.sh
> cat output.tsv | go run cmd/resize/main.go
```
