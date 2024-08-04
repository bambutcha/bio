function scrollToBlock(blockId) {
    const targetBlock = document.getElementById(blockId);
    targetBlock.scrollIntoView({
      behavior: 'smooth',
      block: 'start'
    });
  }