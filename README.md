# l0_golang

Запуск приложения:
```sh
make
```
postgres и kafka запускаются в docker-контейнерах, данные для подключения и запуска сервера указываются в файле "config.yaml", имеется возможность указать порт либо сокет, при пустом конфиге загружаются значения по умолчанию 

Запуск консьюмера:

```sh
make consumer
```

Запуск тестов:
```sh
make test
```

Логи сервера хранятся в logs/server.log
Логи консьюмера kafka хранятся в logs/kafka_consumer.log
