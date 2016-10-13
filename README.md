# HSBeat

Beats for Java HotSpot VM. This beat ships all performance counters in HotSpot VM.


## Features

* HSBeat collects periodically all raw performance counter values in Java HotSpot VM. 
  * Constant values are shipped only once (first time) to Elasticsearch.
  * Monotonic and Variable values are shipped in all collection time.
* If you want to calculate these values (e.g. ratio, time), you have to implement it in your client apps. 
  * HSBeat Kibana dashboard sample use dynamic scripting on Elasticsearch.


## Getting started

```
$ go get github.com/YaSuenag/hsbeat
$ hsbeat -E hsbeat.modules.0.pid=<PID>
```

If you want to use sample dashboard, you can import as below:

```
$ import_dashboards --dir $GOPATH/src/github.com/YaSuenag/hsbeat/etc/kibana
```

```import_dashboards``` is provided by Beats binary. Please see [reference manual](https://www.elastic.co/guide/en/beats/libbeat/5.0/import-dashboards.html) if you want to know more details.

* If you want to use sample dashboard, you have to enable scripting on ```$ES_HOME/config/elasticsearch.yml``` as below:
```
script.engine.groovy.inline: true
```

