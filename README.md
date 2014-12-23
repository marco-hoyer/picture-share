picture-share
=============

A picture downloader written in Go, designed as client for http-based storage like simple apache webserver or s3 f.e.

Just put a json metadata file on your server pointing to available albums, the downloader will get them if not already done.


Serverside metadata.json:

{
    "url": "http://<your-server>/data",
    "albums": [
        {"year": "2013", "name": "Weihnachten und Silvester", "file": "weihnachten_silvester_2013.zip", "users" : ["user1"]}
    ]
}
