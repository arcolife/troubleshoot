## preflight version

Print the current version and exit

### Synopsis

Print the current version and exit

```
preflight version [flags]
```

### Options

```
  -h, --help   help for version
```

### Options inherited from parent commands

```
      --collect-without-permissions   always run preflight checks even if some require permissions that preflight does not have (default true)
      --collector-image string        the full name of the collector image to use
      --collector-pullpolicy string   the pull policy of the collector image
      --debug                         enable debug logging
      --format string                 output format, one of human, json, yaml. only used when interactive is set to false (default "human")
      --interactive                   interactive preflights (default true)
  -o, --output string                 specify the output file path for the preflight checks
      --selector string               selector (label query) to filter remote collection nodes on.
      --since string                  force pod logs collectors to return logs newer than a relative duration like 5s, 2m, or 3h.
      --since-time string             force pod logs collectors to return logs after a specific date (RFC3339)
```

### SEE ALSO

* [preflight](preflight.md)	 - Run and retrieve preflight checks in a cluster

###### Auto generated by spf13/cobra on 21-Nov-2022