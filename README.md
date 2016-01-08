# HSBeat

Beats for Java HotSpot VM.
This beat ships all performance counters in HotSpot VM.

## How to Build

```shell
$ go build
```

## How to use

You should run ```hsbeat``` with same user of target VM.

```shell
$ go get github.com/YaSuenag/hsbeat
$ hsbeat <PID> <Interval (in ms)>
```

You can import Kibana dashboard (etc/kibana.json)

## License
GNU General Public License v2

