# GUI for Mainflux in Elm
Dashboard made with [elm-bootstrap](http://elm-bootstrap.info/).

## Install

### Docker container GUI build

Install Docker (https://docs.docker.com/install/) and Docker compose
(https://docs.docker.com/compose/install/), `cd` to Mainflux root directory and
then

`docker-compose -f docker/docker-compose.yml up`

if you want to launch a whole Mainflux docker composition or just

`docker-compose -f docker/docker-compose.yml up ui`

if you want to launch just GUI.


### Native GUI build

Install Elm (https://guide.elm-lang.org/install.html) and then edit
*src/Env.elm* by replacing

```
env =
    { protocol = ""
    , host = ""
    , port_ = ""
    }
```

with

```
env =
    { protocol = "http"
    , host = "localhost"
    , port_ = "80"
    }
```
So you can send requests with absolute URLs. Put the desired values in the
record fields for the protocol, host and port. Then run the following commands,

```
git clone https://github.com/mainflux/mainflux
cd mainflux/ui
make
```

This will produce `index.html` in the _ui_ directory. In order to use it, `cd`
to _ui_ and do

`make run`

and follow the instructions on screen.

**NB:** `make` does `elm make src/Main.elm` and `make run` just executes `elm
reactor`. Cf. _Makefile_ for more options.

### Contribute to the GUI development

Follow the instructions above to install and run GUI as a native build. Instead
of `make run` you can install `elm-live` (https://github.com/wking-io/elm-live)
and execute `elm-live src/Main.elm` to get a live reload when your `.Elm` pages
change.

Launch Mainflux without ui service, either natively or as a Docker composition.
Follow the guidelines for Mainflux contributors found here
https://mainflux.readthedocs.io/en/latest/CONTRIBUTING/.
