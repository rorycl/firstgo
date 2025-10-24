# firstgo

A web server for prototyping web interfaces using sketches and clickable
zones to move between pages.

![](recording.gif)

## Why

If you are developing a web site or service, perhaps through domain
driven design [techniques](https://en.wikipedia.org/wiki/Event_storming),
starting to put together concept sketches that elucidate the "nouns" and
the "verbs" of a system can be very useful to validate the design. Using
sketches also helps separate the implementation from the all-important
system concepts.

## Howto

1. compile or download the `firstgo` binary
2. sketch some images and save them as jpgs or gifs in `images/`
3. work out where the clickable zones are for each image (gimp works well)
4. record each image and clickable zones in the `config.yaml` file.
5. run with `./firstgo config.yaml` and go to
   [http://127.0.0.1:8000](http://127.0.0.1:8000)

## Configuration

The provided example configuration file sets out the images representing
"pages" and the clickable area on each. Each "Zone" is the top left and
bottom right of a triangle.

```yaml
pageTemplate: "templates/page.html"
pages:
  -
    URL: "/home"
    Title: "Home"
    ImagePath: "images/home.jpg"
    Zones:
      -
        Left:   606
        Top:    33
        Right:  761
        Bottom: 69
        Target: "/about"
      -
        Left:   61
        Top:    202
        Right:  611
        Bottom: 247
        Target: "/detail"
    ...etc...
```

The styling and render templates by be altered by simply editing the
provided files in the `static` and `templates` directories respectively.

## Command

The `firstgo` command has the following options:

```
Usage of ./firstgo:
  -address string
    	server network address (default "127.0.0.1")
  -port string
    	server network port (default "8000")
  <configFile>
    	yaml configuration file

run a "firstgo" webserver to show interactive image "pages".

This programme uses the configuration in config.yaml and images in the
images directory and css in the static directory to serve up interactive
image "web pages" to mock up a web site or service.

eg ./firstgo [-address 192.168.4.5] [-port 8001] <configfile>
```

## Licence

This project is licensed under the [MIT Licence](LICENCE).
