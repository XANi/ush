# USH

µshare - minimalistic file sharing program



## Sharing basic file

Small file sharing program

usage:

    -> ᛯ curl --upload-file 1M http://127.0.0.1:3001/u/filename
    download URL: http://127.0.0.1:3001/d/filename

then to download just use any HTTP client

    -> ᛯ curl http://127.0.0.1:3001/d/filename -o /tmp/asd





### Planned features

* at-rest encryption (password encoded in URL)
* optional E2E encryption (cli client to send already encrypted data)
* multiplexed streaming download (steam files thru server without touching hard drive)
