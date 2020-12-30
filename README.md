# Global Steam

### Processes

Global Steam is split into processes. They are found in the ./cmd folder. To run a process, you can use the helper script:

`./run frontend`, `./run backend` etc

You will need to run `npm install` to install dependencies first.

##### Environment

All configs are handled through environment variables, you can find a list of them all in `config.go`.
You should get warnings if you run a process without a required config set.

### Services

Global Steam uses several third party apps to run. You can install these with Brew:

`brew tap mongodb/brew`

`brew install rabbitmq memcached mongodb-community@4.2 influxdb mysql elasticsearch` 

Or with Docker:

```
version: '3.7'

services:
  rabbit:
    container_name: rabbit
    hostname: rabbit
    image: rabbitmq:latest
    ports:
      - "5672:5672"
    restart: "unless-stopped"
    volumes:
      - ${DATA_DIR}/rabbitmq:/var/lib/rabbitmq
  memcache:
    container_name: memcache
    hostname: memcache
    image: memcached:latest
    restart: "unless-stopped"
    ports:
      - "11211:11211"
  mongo:
    container_name: mongo
    hostname: mongo
    image: mongo:4.2
    restart: "unless-stopped"
    ports:
      - "27017:27017"
    volumes:
      - ${DATA_DIR}/mongodb:/data/db
  influx:
    container_name: influx
    hostname: influx
    image: influxdb:1.7
    restart: "unless-stopped"
    ports:
      - "8086:8086"
    volumes:
      - ${DATA_DIR}/influxdb/data:/root/.influxdb/data
      - ${DATA_DIR}/influxdb/wal:/root/.influxdb/wal
      - ${DATA_DIR}/influxdb/meta:/root/.influxdb/meta
  mysql:
    container_name: mysql
    hostname: mysql
    image: mysql:8.0
    ports:
      - "3306:3306"
    restart: "unless-stopped"
    environment:
      - MYSQL_DATABASE
      - MYSQL_ROOT_PASSWORD
    volumes:
      - ${DATA_DIR}/mysql:/var/lib/mysql
  search:
    container_name: search
    hostname: search
    image: elasticsearch:7.6.2
    ports:
      - "9200:9200"
    restart: "unless-stopped"
    environment:
      - ELASTIC_PASSWORD
    volumes:
      - ${DATA_DIR}/elasticsearch/:/usr/share/elasticsearch/data/
```

### Updating Assets

After updating .sass or .js files you need to compile them by running `npm run webpack`
