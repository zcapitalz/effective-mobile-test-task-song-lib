# Заметки

### Архитектура данных
Т.к. в качестве ID выбран тип данных KSUID, то при пагинации(от меньшего к большему) по такому ID новые элементы попадут в последнюю страницу и не будут пропущены, если они были добавлены после начала пагинации.

### Хранение сущностей
В репозиториях используется билдер запросов для большего контроля за взаимодействием с базой данных. Однако вполне может быть, что в данной задаче лучше подошел бы GORM.