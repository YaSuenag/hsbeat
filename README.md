# HSBeat

Beats for Java HotSpot VM.
This beat ships all performance counters in HotSpot VM.

## How to Build

```shell
$ go get github.com/elastic/libbeat
$ cd src
$ go build hsbeatmain.go
```

## How to use

You should run ```hsbeatmain``` with same user of target VM.

```shell
$ ./hsbeatmain <PID> <Interval (in ms)>
```

You can import Kibana dashboard (etc/kibana.json)

## License
GNU General Public License v2

