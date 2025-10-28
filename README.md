# firstgo

A web server for prototyping web interfaces using sketches and clickable
zones to move between pages.

![](recording.gif)

### Why

If you are developing a web site or service, perhaps through domain
driven design [techniques](https://en.wikipedia.org/wiki/Event_storming),
starting to put together concept sketches that elucidate the "nouns" and
the "verbs" of a system can be very useful to validate the design. Using
sketches also helps separate technical implementation from the
all-important domain concepts.

### Howto

Download the `firstgo` binary for your platform from
[releases](https://github.com/rorycl/firstgo/releases).

`firstgo` runs in `demo`, `init` or `serve` modes:

* **demo**:  
  run the embedded demo to get a feel for how `firstgo` works  
  run `./firstgo demo` 
* **init**:  
  initialise a new project
  run `./firstgo init` to write scaffolding files from the demo project 
* **serve**:  
  serve project files on disk  
  run `./firstgo serve config.yaml`

To deploy your custom content in production, either copy your project
files with the binary to your production setting, or copy your project
yaml and corresponding images, static and templates material to the
`assets` directory and recompile the binary to embed them.

### Configuration & Customisation

The provided example configuration file sets out the images representing
"pages" and the clickable area on each. Each "Zone" is the top left and
bottom right of a rectangle. Notes, in markdown format, can also be added.

See the provided [config.yaml](./config.yaml) for an example.

The styling and render templates may be easily customised by editing the
the css file in `static` and the two [golang
templates](https://www.digitalocean.com/community/tutorials/how-to-use-templates-in-go).
templates in `templates`.

If no pages are configured to be served from `/` and `/index` these
endpoints will be automatically provided with a simple index.

### Utilities

For recording clickable zones on images, consider using a tool like
[LabelImg](https://github.com/HumanSignal/labelImg). Alternatively a
simple bash script using `qiv` and `slop` is provided at
[utils/zone-recorder.sh](utils/zone-recorder.sh). A cross-platform
builder script is also provided.

### Run

The `firstgo` command has the options shown below with `./firstgo -h`.
Invoke a command with `-h` to see its specific help, such as `./firstgo
serve -h`.

```
Usage:
  firstgo 

A web server for prototyping web interfaces using sketches and clickable
zones to move between pages.

 <demo | init | serve>

Help Options:
  -h, --help  Show this help message

Available commands:
  demo   Run the demo server
  init   Init a project
  serve  Serve content on disk with the provided config

```

### Licence

This project is licensed under the [MIT Licence](LICENCE).
