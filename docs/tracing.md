# Tracing

Distributed tracing is a method of profiling and monitoring appplications.  It can provide valuable insight when optimizing and debugging an application.  Mainflux includes the Jaeger open tracing framework as a service with its stack by default.

## Launch

The Jaeger service will launch with the rest of the Mainflux services.  All services can be launched using:

```bash
make run
```

The Jaeger UI can then be accessed at ```http://http://localhost:16686``` from your browser.  Details about the UI can be found in [Jaeger's official documentation](https://www.jaegertracing.io/docs/1.14/frontend-ui/).

## Configure

The Jaeger service can be disabled by using the `scale` flag with ```docker-compose up``` and setting the jaeger container to 0.

```bash
--scale jaeger=0
```

This is currently the only difference when using the ```make rungw``` command versus ```make run```.  
> The ```make rungw``` command runs Mainflux for gateway devices.  There could potentially be more differences running with this command in the future.

Jaeger uses 5 ports within the Mainflux framework.  These ports can be edited within the `.env` file.

| Variable            | Description                                       | Default     |
| ------------------- | ------------------------------------------------- | ----------- |
| MF_JAEGER_PORT      | Agent port for compact jaeger.thrift protocol     | 6831        |
| MF_JAEGER_FRONTEND  | UI port                                           | 16686       |
| MF_JAEGER_COLLECTOR | Collector for jaeger.thrift directly from clients | 14268       |
| MF_JAEGER_CONFIGS   | Configuration server                              | 5778        |
| MF_JAEGER_URL       | Jaeger access from within Mainflux                | jaeger:6831 |

