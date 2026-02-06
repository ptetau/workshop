const button = document.getElementById('ping');
const status = document.getElementById('status');

button.addEventListener('click', () => {
  status.textContent = 'pong';
  status.classList.add('show');
  setTimeout(() => status.classList.remove('show'), 900);
});