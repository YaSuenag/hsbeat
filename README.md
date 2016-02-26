# HSBeat

Beats for Java HotSpot VM.
This beat ships all performance counters in HotSpot VM.

## Features

* HSBeat collects periodically all raw performance counter values in Java HotSpot VM.
 * Constant values are shipped only once (first time) to Elasticsearch.
 * Monotonic and Variable values are shipped in all collection time.
* If you want to calculate these values (e.g. ratio, time), you have to implement it in your client apps.
 * HSBeat Kibana dashboard sample use dynamic scripting on Elasticsearch.

## How to Build

```shell
$ go build
```

## Configuration

You can edit HSBeat configuration in ```hsbeat.yml``` .

* force_collect
 * String array type.
 * Performance counter name which you want to collect in everytime.
  * Delimiter of performance counter name is slash (/).
 * ```"sun/os/hrt/frequency"``` is set by default.
  * This counter value is used in HSBeat Kibana dashboard sample for calculating time values.

## How to use

Run ```hsbeat``` with same user of target VM.

```shell
$ go get github.com/YaSuenag/hsbeat
$ curl -XPUT http://<host>:9200/_template/hsbeat -d@etc/hsbeat-template.json
$ hsbeat -p <PID> -i <Interval (in ms)>
```

Note:
* -p is mandatory, -i is optional (5000 is by default).
* You can see all options with -h.

If you want to use (import) Kibana dashboard sample (etc/kibana.json), you have to enable Elasticsearch dynamic scripting in ```$ES_HOME/config/elasticsearch.yml``` as below:

```
script.inline: true
script.indexed: true
```

See also Elasticsearch reference of [Enabling dynamic scripting](https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-scripting.html#enable-dynamic-scripting) .

## License
GNU General Public License v2

