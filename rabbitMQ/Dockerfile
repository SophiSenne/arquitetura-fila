# Construido com IA

# Imagem oficial com painel de gerenciamento
FROM rabbitmq:3-management

# Definir variáveis padrão (podem ser sobrescritas no docker-compose)
ENV RABBITMQ_DEFAULT_USER=admin
ENV RABBITMQ_DEFAULT_PASS=admin
ENV RABBITMQ_DEFAULT_VHOST=/

# Habilita plugins extras (opcional)
RUN rabbitmq-plugins enable --offline rabbitmq_management

# Expõe portas
# 5672 -> comunicação AMQP
# 15672 -> painel web
EXPOSE 5672 15672

# Diretório de dados para persistência
VOLUME ["/var/lib/rabbitmq"]

# Inicialização
CMD ["rabbitmq-server"]