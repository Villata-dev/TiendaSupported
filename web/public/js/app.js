document.addEventListener('DOMContentLoaded', () => {
    // Referencias a elementos del DOM
    const authSection = document.getElementById('auth-section');
    const productsSection = document.getElementById('products-section');
    const loginForm = document.getElementById('login');
    const registerForm = document.getElementById('register');
    const toggleAuthBtn = document.getElementById('toggle-auth');
    const toggleLoginBtn = document.getElementById('toggle-login');
    const loginDiv = document.getElementById('login-form');
    const registerDiv = document.getElementById('register-form');
    const productList = document.querySelector('editable-list');

    // Función para mostrar mensajes de error/éxito
    const showMessage = (message, isError = false) => {
        alert(message); // Podríamos mejorar esto con un componente de toast/notification
    };

    // Toggle entre login y registro
    toggleAuthBtn?.addEventListener('click', (e) => {
        e.preventDefault();
        loginDiv.style.display = 'none';
        registerDiv.style.display = 'block';
    });

    toggleLoginBtn?.addEventListener('click', (e) => {
        e.preventDefault();
        registerDiv.style.display = 'none';
        loginDiv.style.display = 'block';
    });

    // Función para manejar errores de fetch
    const handleFetchError = async (response) => {
        const data = await response.json().catch(() => ({}));
        if (!response.ok) {
            throw new Error(data.message || 'Error en la operación');
        }
        return data;
    };

    // Manejo de registro
    registerForm?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        try {
            const response = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    username: formData.get('username'),
                    password: formData.get('password')
                })
            });

            await handleFetchError(response);
            showMessage('Registro exitoso. Por favor, inicia sesión.');
            registerDiv.style.display = 'none';
            loginDiv.style.display = 'block';
            e.target.reset();
        } catch (error) {
            showMessage(error.message, true);
            console.error('Error en registro:', error);
        }
    });

    // Manejo de login
    loginForm?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    username: formData.get('username'),
                    password: formData.get('password')
                })
            });

            await handleFetchError(response);
            authSection.style.display = 'none';
            productsSection.style.display = 'block';
            e.target.reset();
            await loadProducts();
        } catch (error) {
            showMessage(error.message, true);
            console.error('Error en login:', error);
        }
    });

    // Cargar productos
    async function loadProducts() {
        try {
            console.log('Intentando cargar productos...');
            const response = await fetch('/api/v1/products');
            console.log('Respuesta del servidor:', response.status, response.statusText);
            
            const data = await handleFetchError(response);
            console.log('Productos cargados:', data);
            productList.setData(data);
        } catch (error) {
            showMessage('Error al cargar productos', true);
            console.error('Error cargando productos:', error);
        }
    }

    // Eventos del Web Component
    productList?.addEventListener('item-create', async (e) => {
        try {
            const response = await fetch('/api/v1/products', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(e.detail)
            });
            await handleFetchError(response);
            await loadProducts();
            showMessage('Producto creado exitosamente');
        } catch (error) {
            showMessage('Error al crear producto', true);
            console.error('Error creando producto:', error);
        }
    });

    productList?.addEventListener('item-delete', async (e) => {
        try {
            const response = await fetch(`/api/v1/products/${e.detail.id}`, {
                method: 'DELETE'
            });
            await handleFetchError(response);
            await loadProducts();
            showMessage('Producto eliminado exitosamente');
        } catch (error) {
            showMessage('Error al eliminar producto', true);
            console.error('Error eliminando producto:', error);
        }
    });

    productList?.addEventListener('item-edit', async (e) => {
        try {
            const response = await fetch(`/api/v1/products/${e.detail.id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(e.detail)
            });
            await handleFetchError(response);
            await loadProducts();
            showMessage('Producto actualizado exitosamente');
        } catch (error) {
            showMessage('Error al actualizar producto', true);
            console.error('Error actualizando producto:', error);
        }
    });
});