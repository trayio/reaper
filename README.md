#### reaper

A distant relative of chaos monkey.

Terminates EC2 instances based on tags and launch time.

```sh
$ ./reaper -h
Usage of ./reaper:
    -c="config.js": Configuration file.
    -dry=false: Enable dry run.
    -tag="group": Tag name to group instances by
```

By default tag used is *group*. Use **-tag** option to select a different tag.


###### Configuration

Configuration sample with two groups defined by tags *nginx-production* and *nginx-staging*.

Group *nginx-production*:

- terminate 3 instances older than 3 days


Group *nginx-staging*:

- terminate 1 instance older than 5 hours

```javascript
{
    "nginx-production": {
        "count": 3,
        "age": 72
    },
    "nginx-staging": {
        "count": 1,
        "age": 5
    }
}
```

Vendor:

- aws-go-sdk v1.1.3
- HCL 1c284ec98f4b398443cbabb0d9197f7f4cc0077c
