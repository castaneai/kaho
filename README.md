kaho
========

![kaho](https://78.media.tumblr.com/cd146dbbde750d75016e8c6bd70fcc0b/tumblr_ozms66YK0D1wqeriwo5_400.png)

## Usage

```sh
curl -si -X POST https://kaho/upload -F "file=@/path/to/image.png;type=image/png" | grep location
```

## Deploy

```sh
gcloud app deploy 
```
