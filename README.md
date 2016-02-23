# HSBeat

Beats for Java HotSpot VM.
This beat ships all performance counters in HotSpot VM.

## How to Build

```shell
$ go build
```

## How to use

Run ```hsbeat``` with same user of target VM.

```shell
$ go get github.com/YaSuenag/hsbeat
$ curl -XPUT http://<host>:9200/_template/hsbeat -d@etc/hsbeat-template.json
$ hsbeat <PID> <Interval (in ms)>
```

If you want to use (import) Kibana dashboard sample (etc/kibana.json), you have to enable Elasticsearch dynamic scripting in ```$ES_HOME/config/elasticsearch.yml``` as below:

```
script.inline: true
script.indexed: true
```

See also Elasticsearch reference of [Enabling dynamic scripting](https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-scripting.html#enable-dynamic-scripting) .

## License
GNU General Public License v2

