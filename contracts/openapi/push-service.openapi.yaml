openapi: 3.0.3
info:
  title: TripGo Push Service API
  version: 1.0.0
  description: >-
    HTTP-контракт сервиса-заглушки пушей. Используется в лабораторной работе 3.
    В лабораторной работе 4 тот же сервис вызывается по gRPC
    (contracts/proto/push/v1/push.proto).
servers:
- url: http://localhost:8090
tags:
- name: Push
- name: Admin
paths:
  /api/v1/push:
    post:
      tags:
      - Push
      summary: Отправить команду на пуш-уведомление
      operationId: sendPush
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PushRequest'
            example:
              request_id: 01J9Z6H7K8QW3M4N5P6R7S8T9V
              recipient_id: 8860b315-ec86-42eb-a17c-7c163d721ff5
              kind: REQUEST_POSITION
              data:
                trip_id: 1f0a9c62-4a1c-4f2e-9d33-2a4bb0f0b111
      responses:
        '202':
          description: Команда принята
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PushAccepted'
        '400':
          description: Некорректный запрос
        '429':
          description: Превышен лимит
          headers:
            Retry-After:
              description: Через сколько секунд повторять
              schema:
                type: integer
        '503':
          description: Временная недоступность
  /admin/behaviour:
    post:
      tags:
      - Admin
      summary: Настроить поведение заглушки на лету
      description: Используется в лабораторной работе 5 для проверки устойчивости.
      operationId: setBehaviour
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Behaviour'
      responses:
        '200':
          description: Настройки применены
        '400':
          description: Некорректные настройки
  /health:
    get:
      tags:
      - Push
      summary: Liveness
      operationId: pushHealth
      responses:
        '200':
          description: Процесс работает
components:
  schemas:
    PushKind:
      type: string
      enum:
      - REQUEST_POSITION
      - TRIP_CREATED
      - TRIP_COMPLETED
    PushRequest:
      type: object
      additionalProperties: false
      required:
      - request_id
      - recipient_id
      - kind
      properties:
        request_id:
          type: string
          description: Идентификатор запроса для корреляции логов.
        recipient_id:
          type: string
          format: uuid
          description: Кому шлём. Для REQUEST_POSITION — driver_id поездки.
        kind:
          $ref: '#/components/schemas/PushKind'
        title:
          type: string
        body:
          type: string
        data:
          type: object
          additionalProperties:
            type: string
    PushAccepted:
      type: object
      additionalProperties: false
      required:
      - message_id
      - accepted_at
      properties:
        message_id:
          type: string
        accepted_at:
          type: string
          format: date-time
    Behaviour:
      type: object
      additionalProperties: false
      properties:
        fail_rate:
          type: number
          format: double
          minimum: 0
          maximum: 1
        latency_ms:
          type: integer
          minimum: 0
        rate_limit_rps:
          type: integer
          minimum: 0
