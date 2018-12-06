kaho
========

![kaho](https://78.media.tumblr.com/cd146dbbde750d75016e8c6bd70fcc0b/tumblr_ozms66YK0D1wqeriwo5_400.png)

## Usage

```sh
curl -X POST https://kaho/upload -F "@/path/to/image.png;type=image/png"
```

## Deploy

```sh
goapp deploy -application [PROJECT_ID] -version [VERSION] appengine/app.yaml
```

## Development

```
dep ensure
goapp serve appengine/app.yaml
```