function scrollToBlock(blockId) {
    const targetBlock = document.getElementById(blockId);
    targetBlock.scrollIntoView({
      behavior: 'smooth',
      block: 'start'
    });
}


const audio = document.getElementById("soundPlayer");

const playButton = document.getElementById("soundButton");

playButton.addEventListener("click", function() {

audio.play();

});

