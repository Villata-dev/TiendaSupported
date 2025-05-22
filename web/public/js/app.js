// TiendaSupported/web/public/js/app.js

// Importar funciones de validación
import { validateProduct } from '/static/js/validators.js'; // Ruta corregida y exportación

// Declarar las funciones en el scope global para que puedan ser llamadas desde el HTML
window.editProduct = null;
window.deleteProduct = null;

document.addEventListener('DOMContentLoaded', async () => {
    // Referencias a elementos del DOM
    const authSection = document.getElementById('auth-section');
    const productsSection = document.getElementById('products-section');
    const loginForm = document.getElementById('login');
    const registerForm = document.getElementById('register');
    const toggleAuthBtn = document.getElementById('toggle-auth');
    const toggleLoginBtn = document.getElementById('toggle-login');
    const loginDiv = document.getElementById('login-form');
    const registerDiv = document.getElementById('register-form');
    const productsTbody = document.getElementById('products-tbody'); // Referencia a la tabla de productos
    const editModal = document.getElementById('edit-modal');
    const editForm = document.getElementById('edit-form');
    const closeModalButtons = document.querySelectorAll('.close-modal');
    const searchInput = document.getElementById('search-input');
    const sortBySelect = document.getElementById('sort-by');
    const stockFilterSelect = document.getElementById('stock-filter');
    const itemsPerPageSelect = document.getElementById('items-per-page');
    const prevPageBtn = document.getElementById('prev-page');
    const nextPageBtn = document.getElementById('next-page');
    const currentPageSpan = document.getElementById('current-page');
    const showingStart = document.getElementById('showing-start');
    const showingEnd = document.getElementById('showing-end');
    const totalItems = document.getElementById('total-items');
    const logoutBtn = document.getElementById('logout-btn');
    const addProductForm = document.getElementById('product-form');


    // Variables para paginación
    let currentPage = 1;
    let itemsPerPage = parseInt(itemsPerPageSelect.value);
    let allProducts = []; // Para mantener la lista completa de productos

    // Función para mostrar mensajes de error/éxito
    const showMessage = (message, isError = false) => {
        const notification = document.createElement('div');
        notification.className = `notification ${isError ? 'error' : 'success'}`;
        notification.textContent = message;
        
        const prevNotification = document.querySelector('.notification');
        if (prevNotification) {
            prevNotification.remove();
        }
        
        document.body.appendChild(notification);
        requestAnimationFrame(() => notification.classList.add('show'));
        
        setTimeout(() => {
            notification.style.transform = 'translateX(120%)';
            notification.addEventListener('transitionend', () => notification.remove());
        }, 3000);
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
        const data = await response.json().catch(() => ({ 
            message: response.status === 409 ? 
                'El usuario ya existe' : 
                'Error de respuesta del servidor.' 
        }));
        if (!response.ok) {
            throw new Error(data.message || `Error ${response.status}: ${response.statusText}`);
        }
        return data;
    };

    // Función de confirmación personalizada (reemplaza alert/confirm)
    const confirmAction = (message) => {
        return new Promise((resolve) => {
            const confirmModal = document.createElement('div');
            confirmModal.className = 'modal confirm-modal show';
            confirmModal.innerHTML = `
                <div class="modal-content">
                    <div class="modal-header">
                        <h2>Confirmar Acción</h2>
                    </div>
                    <div class="modal-body">
                        <p>${message}</p>
                    </div>
                    <div class="modal-footer">
                        <button class="btn btn-secondary" data-action="cancel">Cancelar</button>
                        <button class="btn btn-danger" data-action="confirm">Confirmar</button>
                    </div>
                </div>
            `;

            document.body.appendChild(confirmModal);

            const handleClick = (e) => {
                const action = e.target.dataset.action;
                if (action) {
                    confirmModal.remove();
                    resolve(action === 'confirm');
                }
            };

            confirmModal.addEventListener('click', handleClick);
        });
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

            const userData = await handleFetchError(response); // El backend devuelve datos de usuario
            showMessage('Inicio de sesión exitoso');
            authSection.style.display = 'none';
            productsSection.style.display = 'block';
            e.target.reset();
            await loadProducts();
            // Aquí puedes usar userData.role para mostrar/ocultar elementos si es necesario
            console.log("Usuario logeado:", userData.username, "Rol:", userData.role);

        } catch (error) {
            showMessage(error.message, true);
            console.error('Error en login:', error);
        }
    });

    // Eventos para productos
    logoutBtn?.addEventListener('click', async () => {
        try {
            const confirmed = await confirmAction('¿Estás seguro de que deseas cerrar sesión?');
            if (!confirmed) return;

            const response = await fetch('/api/auth/logout', {
                method: 'POST'
            });
            
            if (response.ok) {
                showMessage('Sesión cerrada exitosamente');
                productsSection.style.display = 'none';
                authSection.style.display = 'block';
                loginDiv.style.display = 'block'; // Mostrar el formulario de login por defecto
            } else {
                const errorData = await response.json();
                showMessage(errorData.message || 'Error al cerrar sesión', true);
            }
        } catch (error) {
            showMessage('Error de conexión al cerrar sesión', true);
            console.error('Error al cerrar sesión:', error);
        }
    });

    addProductForm?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const product = {
            name: formData.get('name'),
            description: formData.get('description'),
            price: parseFloat(formData.get('price')),
            stock: parseInt(formData.get('stock'))
        };

        // Limpiar errores anteriores
        document.querySelectorAll('.error-message').forEach(el => {
            el.textContent = '';
            el.classList.remove('show');
        });
        document.querySelectorAll('input').forEach(el => {
            el.classList.remove('error');
        });

        // Validar datos usando la función importada
        const validation = validateProduct(product);
        if (!validation.isValid) {
            Object.entries(validation.errors).forEach(([field, message]) => {
                const input = document.getElementById(field);
                const errorEl = document.querySelector(`[data-for="${field}"]`);
                
                if (input && errorEl) {
                    input.classList.add('error');
                    errorEl.textContent = message;
                    errorEl.classList.add('show');
                }
            });
            return;
        }

        try {
            const response = await fetch('/api/v1/products', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(product)
            });

            await handleFetchError(response);
            showMessage('Producto agregado exitosamente');
            e.target.reset();
            await loadProducts(); // Recargar productos después de agregar
        } catch (error) {
            showMessage(error.message, true);
            console.error('Error al agregar producto:', error);
        }
    });

    async function loadProducts() {
        try {
            const response = await fetch('/api/v1/products', {
                credentials: 'include' // Incluir credenciales en todas las peticiones
            });
            allProducts = await handleFetchError(response);
            renderProducts(filterProducts(allProducts));
        } catch (error) {
            showMessage(error.message, true);
            console.error('Error cargando productos:', error);
            productsTbody.innerHTML = '<tr><td colspan="6" class="text-center">Error al cargar productos.</td></tr>';
        }
    }

    function filterProducts(products) {
        const searchTerm = searchInput.value.toLowerCase();
        const sortBy = sortBySelect.value;
        const stockFilter = stockFilterSelect.value;

        let filtered = products.filter(product => 
            product.name.toLowerCase().includes(searchTerm) ||
            product.description.toLowerCase().includes(searchTerm)
        );

        if (stockFilter) {
            filtered = filtered.filter(product => {
                switch (stockFilter) {
                    case 'in-stock': return product.stock > 5;
                    case 'low-stock': return product.stock > 0 && product.stock <= 5;
                    case 'out-stock': return product.stock === 0;
                    default: return true;
                }
            });
        }

        if (sortBy) {
            filtered.sort((a, b) => {
                switch (sortBy) {
                    case 'name-asc': return a.name.localeCompare(b.name);
                    case 'name-desc': return b.name.localeCompare(a.name);
                    case 'price-asc': return a.price - b.price;
                    case 'price-desc': return b.price - a.price;
                    case 'stock-asc': return a.stock - b.stock;
                    case 'stock-desc': return b.stock - a.stock;
                    default: return 0;
                }
            });
        }

        return filtered;
    }

    function renderProducts(productsToRender) {
        const startIndex = (currentPage - 1) * itemsPerPage;
        const endIndex = startIndex + itemsPerPage;
        const paginatedProducts = productsToRender.slice(startIndex, endIndex);
        
        // Actualizar información de paginación
        showingStart.textContent = productsToRender.length ? startIndex + 1 : 0;
        showingEnd.textContent = Math.min(endIndex, productsToRender.length);
        totalItems.textContent = productsToRender.length;
        
        // Actualizar estado de botones
        prevPageBtn.disabled = currentPage === 1;
        nextPageBtn.disabled = endIndex >= productsToRender.length;
        currentPageSpan.textContent = `Página ${currentPage}`;

        // Renderizar productos
        productsTbody.innerHTML = paginatedProducts.map((product, index) => `
            <tr style="animation-delay: ${index * 0.05}s">
                <td>${product.id}</td>
                <td>${product.name}</td>
                <td>${product.description}</td>
                <td class="price-column">$${product.price.toFixed(2)}</td>
                <td class="stock-column ${product.stock <= 5 ? 'low-stock' : ''}">${product.stock}</td>
                <td class="actions-column">
                    <div class="btn-group">
                        <button onclick="window.editProduct(${product.id})" class="btn btn-primary btn-sm">
                            Editar
                        </button>
                        <button onclick="window.deleteProduct(${product.id})" class="btn btn-danger btn-sm">
                            Eliminar
                        </button>
                    </div>
                </td>
            </tr>
        `).join('') || '<tr><td colspan="6" class="text-center">No hay productos disponibles</td></tr>';
    }

    // Event listeners para paginación
    itemsPerPageSelect?.addEventListener('change', (e) => {
        itemsPerPage = parseInt(e.target.value);
        currentPage = 1; // Resetear a la primera página al cambiar el número de ítems
        renderProducts(filterProducts(allProducts));
    });

    prevPageBtn?.addEventListener('click', () => {
        if (currentPage > 1) {
            currentPage--;
            renderProducts(filterProducts(allProducts));
        }
    });

    nextPageBtn?.addEventListener('click', () => {
        const filteredProducts = filterProducts(allProducts);
        const totalPages = Math.ceil(filteredProducts.length / itemsPerPage);
        if (currentPage < totalPages) {
            currentPage++;
            renderProducts(filteredProducts);
        }
    });

    // Agregar los event listeners para filtros
    searchInput?.addEventListener('input', () => {
        currentPage = 1; // Resetear a la primera página al cambiar el filtro
        renderProducts(filterProducts(allProducts));
    });

    sortBySelect?.addEventListener('change', () => {
        currentPage = 1; // Resetear a la primera página al cambiar el filtro
        renderProducts(filterProducts(allProducts));
    });

    stockFilterSelect?.addEventListener('change', () => {
        currentPage = 1; // Resetear a la primera página al cambiar el filtro
        renderProducts(filterProducts(allProducts));
    });

    // Asignar las implementaciones a las funciones globales
    window.editProduct = async (id) => {
        try {
            const response = await fetch(`/api/v1/products/${id}`, { credentials: 'include' });
            const product = await handleFetchError(response);
            
            // Llenar el formulario del modal
            document.getElementById('edit-id').value = product.id;
            document.getElementById('edit-name').value = product.name;
            document.getElementById('edit-description').value = product.description;
            document.getElementById('edit-price').value = product.price;
            document.getElementById('edit-stock').value = product.stock;
            
            // Mostrar modal
            editModal.classList.add('show');
        } catch (error) {
            showMessage('Error al cargar el producto para edición', true);
            console.error('Error cargando producto para edición:', error);
        }
    };

    window.deleteProduct = async (id) => {
        try {
            const confirmed = await confirmAction('¿Estás seguro de que deseas eliminar este producto?');
            if (!confirmed) return;

            const response = await fetch(`/api/v1/products/${id}`, {
                method: 'DELETE',
                credentials: 'include'
            });

            await handleFetchError(response);
            showMessage('Producto eliminado exitosamente');
            await loadProducts(); // Recargar productos después de eliminar
        } catch (error) {
            showMessage('Error al eliminar el producto', true);
            console.error('Error eliminando producto:', error);
        }
    };

    // Cerrar modal
    closeModalButtons.forEach(button => {
        button.addEventListener('click', () => {
            editModal.classList.remove('show');
        });
    });

    // Manejo de edición de producto
    editForm?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const id = formData.get('id');

        const updatedProduct = {
            name: formData.get('name'),
            description: formData.get('description'),
            price: parseFloat(formData.get('price')),
            stock: parseInt(formData.get('stock'))
        };

        // Limpiar errores anteriores del modal
        editModal.querySelectorAll('.error-message').forEach(el => {
            el.textContent = '';
            el.classList.remove('show');
        });
        editModal.querySelectorAll('input').forEach(el => {
            el.classList.remove('error');
        });

        const validation = validateProduct(updatedProduct);
        if (!validation.isValid) {
            Object.entries(validation.errors).forEach(([field, message]) => {
                const input = document.getElementById(`edit-${field}`); // Asegúrate de que los IDs coincidan
                const errorEl = editModal.querySelector(`[data-for="${field}"]`);
                
                if (input && errorEl) {
                    input.classList.add('error');
                    errorEl.textContent = message;
                    errorEl.classList.add('show');
                }
            });
            return;
        }

        try {
            const response = await fetch(`/api/v1/products/${id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(updatedProduct),
                credentials: 'include'
            });

            await handleFetchError(response);
            showMessage('Producto actualizado exitosamente');
            editModal.classList.remove('show');
            e.target.reset();
            await loadProducts(); // Recargar productos después de actualizar
        } catch (error) {
            showMessage(error.message, true);
            console.error('Error actualizando producto:', error);
        }
    });

    // Cerrar modal al hacer click fuera de él
    editModal?.addEventListener('click', (e) => {
        if (e.target === editModal) {
            editModal.classList.remove('show');
        }
    });

    // Verificar sesión al cargar la página
    async function checkSession() {
        try {
            const response = await fetch('/api/auth/check-session', {
                credentials: 'include' // Importante: incluir credenciales
            });
            
            if (response.ok) {
                const userData = await response.json(); // Obtener datos de usuario
                console.log("Sesión activa para:", userData.username, "Rol:", userData.role);
                authSection.style.display = 'none';
                productsSection.style.display = 'block';
                await loadProducts();
            } else {
                throw new Error('Sesión inválida o no activa');
            }
        } catch (error) {
            console.error('Error verificando sesión:', error);
            authSection.style.display = 'block';
            productsSection.style.display = 'none';
            loginDiv.style.display = 'block'; // Mostrar el formulario de login por defecto
        }
    }

    // Llamar a checkSession al cargar la página
    await checkSession();
});
