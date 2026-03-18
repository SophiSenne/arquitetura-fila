#Criado com ajuda de IA

FROM rabbitmq:3-management

ENV RABBITMQ_DEFAULT_USER=admin
ENV RABBITMQ_DEFAULT_PASS=admin
ENV RABBITMQ_DEFAULT_VHOST=/

RUN rabbitmq-plugins enable --offline rabbitmq_management

# Copia arquivos de configuração
COPY rabbitmq.conf /etc/rabbitmq/
COPY definitions.json /etc/rabbitmq/

EXPOSE 5672 15672

VOLUME ["/var/lib/rabbitmq"]

CMD ["rabbitmq-server"]