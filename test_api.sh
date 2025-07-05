echo "✅ 1. Создание карточки"
curl -X POST http://localhost:8080/cards \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go","description":"Practice daily","position":1,"images":[]}'
echo -e "\n"

echo "✅ 2. Получение всех карточек"
curl -X GET http://localhost:8080/cards
echo -e "\n"

echo "✅ 3. Обновление позиции карточки (PUT)"
curl -X PUT http://localhost:8080/cards/1 \
  -H "Content-Type: application/json" \
  -d '{"position":5}'
echo -e "\n"

echo "✅ 4. Загрузка изображений"
curl -X POST http://localhost:8080/cards/1/images \
  -H "Content-Type: application/json" \
  -d @images.json
echo -e "\n"

echo "✅ 5. Удаление карточки"
curl -X DELETE http://localhost:8080/cards/1
echo -e "\n"

echo "✅ 6. Проверка, что карточки больше нет"
curl -X GET http://localhost:8080/cards
echo -e "\n"