# HTTP Requester

Simple tool for making _those_ requests to _that_ api and put on _that_ file.

The reason behind this script is that I just grow tired of re-writing the same
ruby script again and again to make calls to whatever api I needed to get
results from a bulk of input.

## Installation

_*Requires go*_

```bash
go get -u github.com/arroadie/http_requester
```
### Usage
```bash
http_requester endpoint[:port] input_file [output_file]
```
If you don't provide a output file, output will go to standard output (so you
may pipe wherever you want to)

### Contribuition

Please

### License

> Copyright © 2016 [Thiago Costa](mailto:thiago@arroadie.com)
> This work is free. You can redistribute it and/or modify it under the
> terms of the Do What The Fuck You Want To Public License, Version 2,
> as published by Sam Hocevar. See the COPYING file for more details.
