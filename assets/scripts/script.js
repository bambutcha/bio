// Функция для прокрутки к указанной секции
function scrollToSection(sectionId) {
  const section = document.getElementById(sectionId);
  if (section) {
    window.scrollTo({
      top: section.offsetTop - 70,
      behavior: 'smooth'
    });
  }
}

// Функция для отображения уведомлений
function showNotification(message, isSuccess = true) {
  // Проверяем, существует ли уже элемент уведомления
  let notification = document.getElementById('notification');
  
  // Если нет, создаем его
  if (!notification) {
    notification = document.createElement('div');
    notification.id = 'notification';
    notification.className = 'notification hidden';
    
    const notificationContent = document.createElement('div');
    notificationContent.className = 'notification-content';
    
    const notificationMessage = document.createElement('span');
    notificationMessage.id = 'notification-message';
    
    const closeButton = document.createElement('button');
    closeButton.className = 'notification-close';
    closeButton.innerHTML = '&times;';
    closeButton.addEventListener('click', function() {
      notification.classList.add('hidden');
    });
    
    notificationContent.appendChild(notificationMessage);
    notificationContent.appendChild(closeButton);
    notification.appendChild(notificationContent);
    
    document.body.appendChild(notification);
  }
  
  // Получаем элемент сообщения
  const notificationMessage = document.getElementById('notification-message');
  
  // Обновляем классы и текст
  notification.classList.remove('hidden', 'success', 'error');
  notification.classList.add(isSuccess ? 'success' : 'error');
  notificationMessage.textContent = message;
  
  // Автоматическое закрытие через 5 секунд
  setTimeout(() => {
    notification.classList.add('hidden');
  }, 5000);
}

// Дождаемся полной загрузки DOM перед выполнением скрипта
document.addEventListener('DOMContentLoaded', function() {
  console.log('DOM полностью загружен');
  
  // "Пробуждение" бэкенда при загрузке страницы
  if (document.getElementById('contactForm')) {
    fetch('https://bambutcha-contact-form.onrender.com/ping')
      .then(response => response.json())
      .then(data => console.log('Backend active, time:', data.time))
      .catch(error => console.log('Ping error:', error));
  }
  
  // Инициализация аудиоплеера
  const audio = document.getElementById('notif');
  const soundButton = document.getElementById('soundButton');
  
  if (audio && soundButton) {
    console.log('Звуковые элементы найдены');
    let isPlaying = false;
    
    // Убедимся, что аудио загружено
    audio.load();
    
    soundButton.addEventListener('click', function() {
      console.log('Клик по звуковой кнопке');
      if (!isPlaying) {
        console.log('Попытка воспроизведения');
        // Принудительно устанавливаем громкость
        audio.volume = 1.0;
        
        // Проверяем состояние аудио
        console.log('Состояние аудио:', audio.readyState);
        
        // Пробуем воспроизвести с обработкой обещания
        let playPromise = audio.play();
        
        if (playPromise !== undefined) {
          playPromise.then(() => {
            console.log('Аудио успешно воспроизводится');
            soundButton.classList.add('play');
            isPlaying = true;
          }).catch(error => {
            console.error('Ошибка при воспроизведении:', error);
            // Попробуем еще раз с задержкой
            setTimeout(() => {
              audio.play().catch(e => console.error('Повторная ошибка:', e));
            }, 1000);
          });
        }
      } else {
        console.log('Останавливаем воспроизведение');
        audio.pause();
        audio.currentTime = 0;
        soundButton.classList.remove('play');
        isPlaying = false;
      }
    });
    
    audio.addEventListener('ended', function() {
      console.log('Воспроизведение завершено');
      soundButton.classList.remove('play');
      isPlaying = false;
    });
  } else {
    console.error('Не найдены элементы аудио или кнопка:', { audio, soundButton });
  }
  
  // Переключение темы
  const themeToggle = document.getElementById('theme-toggle');
  const themeSwitch = document.querySelector('.theme-switch');
  
  if (themeSwitch && themeToggle) {
    console.log('Элементы переключения темы найдены');
    
    // Добавляем обработчик на весь контейнер кнопки
    themeSwitch.addEventListener('click', function() {
      console.log('Клик по кнопке переключения темы');
      const body = document.body;
      body.classList.toggle('dark-theme');
      
      if (body.classList.contains('dark-theme')) {
        themeToggle.classList.remove('fa-moon');
        themeToggle.classList.add('fa-sun');
        localStorage.setItem('theme', 'dark');
      } else {
        themeToggle.classList.remove('fa-sun');
        themeToggle.classList.add('fa-moon');
        localStorage.setItem('theme', 'light');
      }
    });
    
    // Проверяем сохраненную тему
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'dark') {
      document.body.classList.add('dark-theme');
      themeToggle.classList.remove('fa-moon');
      themeToggle.classList.add('fa-sun');
    }
  } else {
    console.error('Элементы переключения темы не найдены', { themeToggle, themeSwitch });
  }

  
  // Мобильное меню
  const hamburger = document.querySelector('.hamburger');
  const navLinks = document.querySelector('.nav-links');
  
  if (hamburger && navLinks) {
    hamburger.addEventListener('click', function() {
      hamburger.classList.toggle('active');
      navLinks.classList.toggle('active');
    });
    
    const navItems = document.querySelectorAll('.nav-links a');
    navItems.forEach(item => {
      item.addEventListener('click', function() {
        hamburger.classList.remove('active');
        navLinks.classList.remove('active');
      });
    });
  }
  
  // Форма обратной связи
  const contactForm = document.getElementById('contactForm');
  if (contactForm) {
    let isFirstRequest = true;
    let timeoutId;
    
    contactForm.addEventListener('submit', async function(e) {
      e.preventDefault();
      
      // Показываем индикатор загрузки
      const submitButton = contactForm.querySelector('button[type="submit"]');
      const originalText = submitButton.textContent;
      submitButton.textContent = 'Отправка...';
      submitButton.disabled = true;
      
      // Если это может быть первый запрос после "сна"
      if (isFirstRequest) {
        // Показываем уведомление о возможной задержке через 2 секунды
        timeoutId = setTimeout(() => {
          showNotification('Первая отправка может занять до 30 секунд...', 'info');
        }, 2000);
      }
      
      try {
        const startTime = new Date().getTime();
        
        const response = await fetch('https://bambutcha-contact-form.onrender.com/api/contact', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            name: document.getElementById('name').value,
            email: document.getElementById('email').value,
            subject: document.getElementById('subject')?.value || '',
            message: document.getElementById('message').value
          })
        });
        
        const endTime = new Date().getTime();
        const requestTime = endTime - startTime;
        
        // Отключаем уведомление о задержке
        clearTimeout(timeoutId);
        
        // Если запрос занял менее 2 секунд, значит сервис не "спал"
        isFirstRequest = requestTime > 2000;
        
        const result = await response.json();
        
        if (result.success) {
          showNotification('Спасибо за сообщение! Я свяжусь с вами в ближайшее время.', true);
          contactForm.reset();
        } else {
          showNotification(result.message || 'Ошибка отправки сообщения', false);
        }
      } catch (error) {
        clearTimeout(timeoutId);
        console.error('Ошибка:', error);
        showNotification('Произошла ошибка при отправке сообщения. Пожалуйста, попробуйте снова позже.', false);
      } finally {
        // Восстанавливаем состояние кнопки
        submitButton.textContent = originalText;
        submitButton.disabled = false;
      }
    });
  }
  
  // Кнопка возврата наверх
  const backToTop = document.getElementById('back-to-top');
  if (backToTop) {
    backToTop.addEventListener('click', function() {
      window.scrollTo({
        top: 0,
        behavior: 'smooth'
      });
    });
  }
});

// Обработчик прокрутки страницы
window.addEventListener('scroll', function() {
  // Фиксированная навигация при прокрутке
  const navbar = document.querySelector('.navbar');
  const logo = document.querySelector('.logo');
  if (navbar || logo) {
    if (window.scrollY > 50) {
      logo.classList.add('scrolledLogo');
      navbar.classList.add('scrolled');
      console.log('Навигация прокручена, добавлен класс scrolled');
    } else {
      logo.classList.remove('scrolledLogo');
      navbar.classList.remove('scrolled');
      console.log('Навигация в верхнем положении, удален класс scrolled');
    }
  }
  
  // Кнопка наверх
  const backToTop = document.getElementById('back-to-top');
  if (backToTop) {
    if (window.scrollY > 300) {
      backToTop.classList.add('show');
    } else {
      backToTop.classList.remove('show');
    }
  }
  
  // Анимация при прокрутке
  const elements = document.querySelectorAll('.service-card, .project-card, .skill-badge');
  elements.forEach(element => {
    const position = element.getBoundingClientRect();
    // Если элемент в области видимости
    if (position.top < window.innerHeight * 0.9) {
      element.classList.add('animate');
    }
  });
});