document.addEventListener('DOMContentLoaded', () => {
    const notificationSystem = {
        show: (message, isError = false) => {
            const notification = document.createElement('div');
            notification.className = `notification ${isError ? 'error' : ''}`;
            notification.textContent = message;
            document.body.appendChild(notification);
            setTimeout(() => notification.classList.add('show'), 100);
            setTimeout(() => {
                notification.classList.remove('show');
                setTimeout(() => notification.remove(), 300);
            }, 3000);
        }
    };

    // Exponer el sistema de notificaciones globalmente
    window.showNotification = notificationSystem.show;
});