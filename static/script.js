const api = "http://localhost:8080/cards";

async function fetchCards() {
  const res = await fetch(api);
  const cards = await res.json();
  const container = document.getElementById("cards");
  container.innerHTML = "";
  cards.forEach(card => {
    const div = document.createElement("div");
    div.className = "card";
    div.innerHTML = `
      <div class="card-content">
        <h3>${card.title}</h3>
        <p>${card.description}</p>
        <p><strong>Позиция:</strong> <input type="number" value="${card.position}" onchange="updatePosition(${card.id}, this.value)"></p>
        <div class="images">
          ${card.images?.map(img => `<img src="${img}" alt="image">`).join("") || ""}
        </div>
        <button onclick="deleteCard(${card.id})">Удалить</button>
      </div>
    `;
    container.appendChild(div);
  });
}

async function createCard() {
  const title = document.getElementById("title").value;
  const description = document.getElementById("description").value;
  const position = parseInt(document.getElementById("position").value);
  const images = document.getElementById("images").value.split(",").map(s => s.trim());

  const res = await fetch(api, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title, description, position, images })
  });

  if (res.ok) {
    const newCard = await res.json();
    if (images.length > 0) {
      await fetch(`${api}/${newCard.id}/images`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(images)
      });
      setTimeout(fetchCards, 1000);
    } else {
      fetchCards();
    }
    document.getElementById("addForm").reset();
  } else {
    alert("Ошибка при создании карточки");
  }
}

async function deleteCard(id) {
  const res = await fetch(`${api}/${id}`, { method: "DELETE" });
  if (res.ok) {
    fetchCards();
  } else {
    alert("Ошибка при удалении");
  }
}

async function updatePosition(id, pos) {
  const res = await fetch(`${api}/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ position: parseInt(pos) })
  });

  if (!res.ok) {
    alert("Ошибка при обновлении позиции");
  }
}

fetchCards();
