# arquitetura-fila



Enunciado:
Uma empresa de monitoramento industrial está modernizando parte de sua operação por meio da adoção de dispositivos embarcados capazes de coletar dados de sensores distribuídos em diferentes ambientes. Esses dispositivos são responsáveis por capturar leituras periódicas de variáveis físicas relevantes para o negócio, como temperatura, umidade, presença, vibração, luminosidade e nível de reservatórios. Parte desses sensores gera valores discretos, como ligado/desligado, presença/ausência ou aberto/fechado, enquanto outros produzem valores analógicos, representando medições contínuas em escalas numéricas.
Com o crescimento do número de dispositivos conectados, a empresa passa a enfrentar desafios relacionados à escalabilidade, confiabilidade e desempenho do backend responsável por receber essas informações. O sistema deve ser capaz de lidar com um grande volume de requisições simultâneas, garantindo que nenhum dado seja perdido, mesmo em cenários de alta concorrência. Além disso, o processamento síncrono direto no momento da requisição pode gerar gargalos e comprometer a estabilidade da aplicação. Diante disso, torna-se necessário adotar uma arquitetura desacoplada, baseada em mensageria, que permita absorver picos de carga e garantir maior resiliência.
Nesse contexto, a solução proposta consiste no desenvolvimento de um backend em GoLang que exponha um endpoint HTTP do tipo POST, responsável por receber pacotes de telemetria enviados pelos dispositivos embarcados. Cada requisição deve conter informações como identificação do dispositivo, timestamp, tipo do sensor, natureza da leitura (discreta ou analógica) e o valor coletado. Após o recebimento, os dados não devem ser processados diretamente no endpoint, mas sim encaminhados para uma fila no RabbitMQ, permitindo que consumidores realizem o processamento de forma assíncrona.
Como parte do fluxo de processamento, os dados consumidos da fila devem ser persistidos em um banco de dados relacional, cuja escolha fica a critério do estudante, sendo fortemente recomendado o uso do PostgreSQL. O modelo de dados deve ser projetado de forma a suportar diferentes tipos de sensores e leituras, garantindo consistência e possibilitando futuras consultas analíticas. O serviço de banco de dados também deve estar integrado à infraestrutura da solução.
Toda a infraestrutura da aplicação deve ser conteinerizada, garantindo portabilidade, reprodutibilidade e facilidade de execução do ambiente. Espera-se que os estudantes utilizem ferramentas como Docker (e, opcionalmente, Docker Compose) para orquestrar os serviços necessários, incluindo o backend, o broker de mensageria (RabbitMQ) e o banco de dados relacional. O repositório do projeto deve conter todos os artefatos necessários para subir o ambiente de forma automatizada, incluindo arquivos de configuração, instruções claras de execução e eventuais scripts auxiliares.
Além disso, será necessário realizar um teste de carga utilizando o k6, simulando múltiplos dispositivos embarcados enviando requisições simultaneamente ao sistema. O objetivo é avaliar o comportamento da aplicação sob diferentes níveis de estresse, analisando métricas como throughput, latência, taxa de erro e capacidade de enfileiramento. Como evidência da avaliação experimental, o repositório deverá incluir relatórios de execução da plataforma, contendo os resultados dos testes realizados, bem como uma análise interpretativa dos dados obtidos, discutindo possíveis gargalos e melhorias.


Arquitetura:
container back (golang) <-> Container RabbitMQ <-> Container Consumidor <-> Banco Relacional

Entregas:
- readme
- dockercompose
- load test
- testes unitarios




[ ] testes de carga
[ ] documentar api
[ ] documentar banco
[ ] documentar arquitetura