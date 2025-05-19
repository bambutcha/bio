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

// Дождаемся полной загрузки DOM перед выполнением скрипта
document.addEventListener('DOMContentLoaded', function() {
  console.log('DOM полностью загружен');
  
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
    contactForm.addEventListener('submit', function(e) {
      e.preventDefault();
      alert('Спасибо за сообщение! В реальной версии сайта оно было бы отправлено.');
      contactForm.reset();
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