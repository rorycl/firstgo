# firstgo

A web server for prototyping web interfaces using sketches and clickable
zones to move between pages.

![](recording.gif)

## Why

If you are developing a web site or service, perhaps through domain
driven design [techniques](https://en.wikipedia.org/wiki/Event_storming),
starting to put together concept sketches that elucidate the "nouns" and
the "verbs" of a system can be very useful to validate the design. Using
sketches also helps separate technical implementation from the
all-important domain concepts.

## Howto

Download the `firstgo` binary for your platform from
[releases](https://github.com/rorycl/firstgo/releases).

`firstgo` runs in `demo`, `init`, `serve` or `develop` modes:

* **demo**: `./firstgo demo` runs the embedded demo to show how
  `firstgo` works
* **init**: `./firstgo init` initialises a new project by writing the
  demo project to disk
* **serve**: `./firstgo serve config.yaml` serves project files from
  disk
* **develop**: `./firstgo develop config.yaml` serves project files from
  disk with automatic reloads of the yaml and template files.

To deploy your custom content in production, either copy your project
files with the binary to your production setting, or copy your project
yaml and corresponding images, static and templates material to the
`assets` directory and recompile the binary to embed them.

## Configuration & Customisation

The configuration file sets out the images representing "pages" and the
clickable area on each. Each "Zone" is the top left and bottom right of
a rectangle. Notes can also be added in markdown format. See the
provided [config.yaml](./config.yaml) for an example.

The styling and render templates can be easily customised by editing the
the css file in `static` and the two [golang
templates](https://www.digitalocean.com/community/tutorials/how-to-use-templates-in-go).
templates in `templates`.

If no pages are configured to be served from `/` and `/index` these
endpoints will be automatically provided with a simple index.

## Record clickable zones

Information on recording clickable zones, including a handy script, is
set out in [utils](./utils/).

## Run

```
./firstgo -h

NAME:
   firstgo - A web server for prototyping web interfaces from sketches

USAGE:
   firstgo [global options] [command]

DESCRIPTION:
   The firstgo server uses a config.yaml file to describe clickable
   zones on images in assets/images to build an interactive website.
   
   For a demo with embedded assets and config file, use 'demo'.
   To start a new project, use 'init' to write the demo files to disk.
   To serve files on disk use 'serve'.
   To serve files on disk in development mode use 'develop'.

COMMANDS:
   serve    Serve content on disk
   develop  Serve content on disk with automatic file reloads
   init     Initialize a new project in a directory
   demo     Run the embedded demo server
   help     Shows a list of commands or help for one command

Run 'firstgo [command] --help' for more information on a command.
```

## Licence

This project is licensed under the [MIT Licence](LICENCE).
