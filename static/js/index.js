document.querySelector('.input-form').addEventListener('submit', async (e) => {
    e.preventDefault();

    const id = document.querySelector('.input').value.trim();
    if (!id) return alert("Введите ID");

    const res = await fetch(`/order/${id}`);
    if (!res.ok) return alert("Заказ не найден");

    const order = await res.json();
    
    // Очищаем предыдущие данные (если были)
    const resultContainer = document.querySelector('.result-container');
    if (resultContainer) resultContainer.remove();

    // Создаем контейнер для вывода данных
    const container = document.createElement('div');
    container.className = 'result-container';
    
    // Форматируем JSON в читаемый вид
    const formattedData = JSON.stringify(order, null, 2);
    container.innerHTML = `<pre>${formattedData}</pre>`;
    
    // Добавляем контейнер после формы
    document.querySelector('.wrapper').appendChild(container);
});