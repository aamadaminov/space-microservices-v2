# space microservices
<img width="1693" height="1241" alt="omicron" src="https://github.com/user-attachments/assets/dbc79bd8-59da-4894-ad66-171042b87f6b" />

Пояснение к узлам на схеме:
1. gencoords - генератор координат от текущих даты/времени.
2. gentemps - генератор температуры/влажности в отсеках корабля от текущих даты/времени.
3. producercoords - продюсер данных в Kafka от генератора координат.
4. producertemps - продюсер данных в Kafka от генератора температуры/влажности.
5. consumersensors - консьюмер данных из Kafka. Сохраняет данные из двух топиков по двум соответствующим таблицам в БД ClickHouse.
6. genflightbook - генератор сообщений из журнала от текущих даты/времени.
7. producerflightbook - продюсер данных в RabbitMQ от генератора собщений.
8. consumernotes - консьюмер данных из RabbitMQ в PostgreSQL с одновременным кэшированием в Redis. Проверить кеширование можно через двукратный HTTP Query (*/api/q?id=*) и Jaeger по трассировкам.
9. genimages - генератор изображений от 10 датчиков. Складывает изображения по папкам.
10. consumerimages - консьюмер сохраняющий изображения от 10 датчиков в хранилище MinIO (S3) с присвоением тегов изображениям. 
11. Elasticsearch + Kibana - сбор и визуализация логов со всех узлов с помощью размещенного на каждом из них ПО FileBeat.
12. Opentelemetry Collector - сбор и перенаправление трейсов и метрик (с их именованием) от приложений и узлов для Jaeger и Prometheus.
13. Prometheus - сбор метрик с Opentelemetry Collector для последующей визуализации в Grafana.
14. Jaeger - сбор трейсов с Opentelemetry Collector и их визуализация.  

Используемые продукты и технологии:
- Clickhouse (clickhouse/clickhouse-server:25.10.2)
- PostgreSQL (postgres:15)
- Redis (redis:7)
- MinIO (minio/minio:RELEASE.2025-09-07T16-13-09Z)
- Kafka (confluentinc/cp-kafka:7.9.5)
- Kafka Exporter (danielqsj/kafka-exporter)
- Zookeeper (confluentinc/cp-zookeeper:7.9.5)
- RabbitMQ (rabbitmq:4.2.0-management)
- Elasticsearch (elastic/elasticsearch:8.17.10)
- Kibana (elastic/kibana:8.17.10)
- Opentelemetry Collector (otel/opentelemetry-collector:0.141.0-amd64)
- Prometheus (prom/prometheus:v3.8.0)
- Grafana (grafana/grafana:12.1.4)
- Jaeger (jaegertracing/jaeger:2.13.0)
- FileBeat (elastic/filebeat:7.17.26)
- Prometheus Node Exporter (prom/node-exporter:v1.10.2)
- gRPC (обмен данными между генераторами данных и продюсерами), HTTP Query (для изменения скорости запросов от продюсеров к генераторам данных)
- otlptrace, otel/exporters/prometheus, minio-go, Sarama (Go-клиент для Apache Kafka)  

Порядок развёртывания (без Kubernetes):
1. В облаке (например, Yandex Cloud) регистрируем в аренду несколько виртуальных машин, достаточно будет 2 ядра Zen 4, 1-6 GB RAM, 10-20 GB SSD.
2. Устанавливаем на хосты Docker, опционально Samba для прямого переноса данных.
3. Запускаем файлы Docker Compose (размещены по папкам с наименованием хостов в каталоге config).
4. Собираем контейнеры как удобно: например, через GitHub Actions с последующей публикацией в ghcr.io, Yandex Cloud Container Registry или же собираем на месте из исходников образы на базе alpine:3.22.2.
