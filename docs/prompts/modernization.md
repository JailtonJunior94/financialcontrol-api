# Modernização: Monolito Modular e Clean Arch

Este prompt foi projetado para guiar a reestruturação do projeto para uma arquitetura de Monolito Modular, garantindo baixo acoplamento e alta coesão entre os domínios de negócio.

## Prompt para IA

"Atue como um arquiteto de software sênior especializado em Go e sistemas distribuídos. O objetivo é criar um plano detalhado para modernizar o projeto 'financialcontrol-api' seguindo o padrão de Monolito Modular e Clean Architecture (Ports and Adapters).

Atualmente, o projeto possui diversos módulos em `internal/modules/` (como billing, cards, catalog). A comunicação entre esses módulos deve ser desacoplada através de eventos em runtime, utilizando o `internal/platform/events/dispatcher.go` já existente.

### Sua tarefa deve incluir:

1.  **Definição da Estrutura de Módulos:** Como cada módulo deve ser estruturado internamente (Domain/Entities, Application/UseCases, Infrastructure/Adapters) para garantir isolamento total da lógica de negócio.
2.  **Estratégia de Eventos:** Como definir Domain Events dentro de cada módulo e como registrá-los no `Dispatcher` central sem criar acoplamento cíclico entre pacotes.
3.  **Isolamento de Persistência:** Definir regras para que cada módulo gerencie seu próprio esquema ou tabelas, garantindo que um módulo não acesse o banco de dados de outro diretamente.
4.  **Injeção de Dependências:** Como organizar o bootstrap da aplicação em `internal/bootstrap` para injetar o `Dispatcher` nos módulos e registrar os listeners necessários.
5.  **Passos de Migração:** Um roadmap de 5 etapas incrementais para transformar o código atual na arquitetura alvo, começando pelo desacoplamento do módulo de `transactions`.

### Critérios de Aceite:
- Proibir explicitamente imports cruzados entre as camadas de `Domain` dos módulos.
- Utilizar o `Dispatcher` síncrono atual, mas preparar o design para uma futura migração para um Message Broker (como Kafka ou RabbitMQ).
- Exemplos de código Go devem ser idiomáticos e seguir os padrões do repositório.

### Saída Esperada:
Um documento Markdown técnico com explicações arquiteturais e exemplos práticos de código para o fluxo de um evento entre dois módulos distintos."
