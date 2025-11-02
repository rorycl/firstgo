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
   firstgo [command] [options]

DESCRIPTION:
   The firstgo server uses a configuration yaml file with templates in
   assets/templates and css in assets/static to describe clickable zones
   on images in assets/images to create an interactive website.

COMMANDS:
   demo     Run the demo server with embedded assets
   init     Initialize a new project from the embedded demo assets
   serve    Serve content on disk
   develop  Serve content on disk with automatic file reloads
   help     Shows a list of commands or help for one command

Run 'firstgo [command] --help' for more information on a command.
```

## Licence

This project is licensed under the [MIT Licence](LICENCE).
