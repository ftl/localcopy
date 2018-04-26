# localcopy
The library localcopy manages a local copy of a resource that is available through HTTP(s). It provides functions do download the resource and to check if an update of the local copy is necessary. 

The update check is done using a HEAD request and comparing the last modification date of the local copy with the Last-Modified header of the HTTP response.

This tool is written in Go on Linux. It might also work on OS X or Windows, but I did not try that out.

## Disclaimer
I develop this tool for myself and just for fun in my free time. If you find it useful, I'm happy to hear about that. If you have trouble using it, you have all the source code to fix the problem yourself (although pull requests are welcome).

## License
This tool is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)
