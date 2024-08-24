function scrollToBlock(blockId) {
  const targetBlock = document.getElementById(blockId);
  targetBlock.scrollIntoView({
    behavior: 'smooth',
    block: 'start'
  });
}

document.addEventListener('DOMContentLoaded', function() {
  var audio = document.getElementById('notif');
  var soundButton = document.getElementById('soundButton');
  var clickCount = 0;

  soundButton.addEventListener('click', function() {
    clickCount++;
    if (clickCount === 1) {
      audio.play();
      soundButton.classList.add('play');
    } else if (clickCount === 2) {
      audio.pause();
      soundButton.classList.remove('play');
      audio.currentTime = 0;
    } else if (clickCount === 3) {
      audio.currentTime = 0;
      audio.play();
      soundButton.classList.add('play');
      clickCount = 1;
    }
  });

  audio.addEventListener('ended', function() {
    soundButton.classList.remove('play');
    clickCount = 0; // Сбрасываем счетчик нажатий
  });
});

