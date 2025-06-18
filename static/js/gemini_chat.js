document.addEventListener('DOMContentLoaded', () => {
  const STORAGE_KEY = 'chatHistory';
  let chatHistory = JSON.parse(sessionStorage.getItem(STORAGE_KEY) || '[]');

  const container = document.getElementById('chat-container');
  const form      = document.getElementById('chat-form');
  const input     = document.getElementById('message');
  const newConv   = document.getElementById('new-conversation-btn');

  function saveHistory() {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(chatHistory));
  }

  function renderBubble(text, role, save = true) {
    const msgContainer = document.createElement('div');
    msgContainer.className = `chat-message ${role}`;

    const msgContent = document.createElement('div');
    msgContent.className = 'chat-message-content';
    
    let cleanedText = text.replace(/[ \t]+/g, ' '); 
    cleanedText = cleanedText.split('\n').map(line => line.trim()).join('\n'); 
    cleanedText = cleanedText.replace(/\n\n\n+/g, '\n\n');
    cleanedText = cleanedText.trim(); 

    let htmlContent = cleanedText
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>') 
      .replace(/\*(.*?)\*/g, '<em>$1</em>');       

    if (htmlContent.includes('\n\n')) {
        htmlContent = htmlContent.split(/\n\n/).map(paragraph => { 
            if (paragraph.trim() === '') return ''; 
            return '<p>' + paragraph.replace(/\n/g, '<br>') + '</p>'; 
        }).join('');
    } else { 
        if (htmlContent.trim() !== '') {
            htmlContent = '<p>' + htmlContent.replace(/\n/g, '<br>') + '</p>';
        } else {
            htmlContent = ''; 
        }
    }
    
    if (htmlContent.trim() === '') {
        htmlContent = '<p>&nbsp;</p>';
    }
    
    msgContent.innerHTML = htmlContent;


    msgContainer.appendChild(msgContent);
    
    container.append(msgContainer);
    container.scrollTop = container.scrollHeight;

    if (save) {
      chatHistory.push({ text, role }); 
      saveHistory();
    }
  }

  function showLoading() {
    const skeletonContainer = document.createElement('div');

    skeletonContainer.className = 'chat-message bot skeleton'; 
    
    const skeletonContent = document.createElement('div');
    skeletonContent.className = 'chat-message-content'; 
    skeletonContainer.appendChild(skeletonContent);

    container.append(skeletonContainer);
    container.scrollTop = container.scrollHeight;
    return skeletonContainer; 
  }

  function loadHistory() {
    container.innerHTML = ''; 
    chatHistory.forEach(msg => {
        if (msg.text && msg.role) { 
             renderBubble(msg.text, msg.role, false);
        }
    });
  }

  newConv.addEventListener('click', () => {
    sessionStorage.removeItem(STORAGE_KEY); 
    chatHistory = [];
    container.innerHTML = '';
    if(input) input.focus();
  });

  form.addEventListener('submit', async e => {
    e.preventDefault();
    const text = input.value.trim();
    if (!text) return;

    renderBubble(text, 'user', true); 
    input.value = '';
    if(input) input.focus();

    const loader = showLoading();
    try {

      const res = await fetch('/api/gemini-chat', {
        method: 'POST',
        headers: {'Content-Type':'application/json'},
        body: JSON.stringify({ 
            message: text, 
        })
      });
      
      if (loader) loader.remove();

      if (!res.ok) {
        const errText = await res.text();
        renderBubble(`Erreur ${res.status} : ${errText || 'Erreur inconnue du serveur'}`, 'bot');
        return;
      }
      const { reply } = await res.json();
      renderBubble(reply || 'Désolé, je n\'ai pas de réponse pour le moment.', 'bot', true);
    } catch (err) {
      if (loader) loader.remove();
      renderBubble(`Erreur de communication avec le serveur : ${err.message}`, 'bot');
    }
  });

  loadHistory(); 
  if(input) input.focus();
});