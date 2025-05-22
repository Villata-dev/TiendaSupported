class EditableList extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.products = [];
    }

    static get styles() {
        return `
            :host {
                --primary-color: #3B82F6;
                --primary-hover: #2563EB;
                --error-color: #EF4444;
                --border-radius: 0.5rem;
                display: block;
                font-family: 'Inter', system-ui, -apple-system, sans-serif;
            }
            
            .product-list {
                width: 100%;
            }
            
            table {
                width: 100%;
                border-collapse: separate;
                border-spacing: 0;
                margin-top: 1.5rem;
            }
            
            th, td {
                padding: 1rem;
                text-align: left;
                border-bottom: 1px solid #E5E7EB;
            }
            
            th {
                background-color: #F9FAFB;
                font-weight: 500;
                color: #4B5563;
                font-size: 0.875rem;
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }
            
            tr:hover td {
                background-color: #F3F4F6;
            }
            
            .actions {
                display: flex;
                gap: 0.5rem;
            }
            
            button {
                padding: 0.5rem 1rem;
                border: none;
                border-radius: var(--border-radius);
                cursor: pointer;
                font-size: 0.875rem;
                font-weight: 500;
                transition: all 0.15s ease;
            }
            
            .edit-btn {
                background-color: var(--primary-color);
                color: white;
            }
            
            .edit-btn:hover {
                background-color: var(--primary-hover);
            }
            
            .delete-btn {
                background-color: var(--error-color);
                color: white;
            }
            
            .delete-btn:hover {
                opacity: 0.9;
            }
            
            form {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                gap: 1rem;
                margin: 1.5rem 0;
                padding: 1.5rem;
                background-color: #F9FAFB;
                border-radius: var(--border-radius);
            }
            
            input {
                padding: 0.75rem 1rem;
                border: 1px solid #E5E7EB;
                border-radius: var(--border-radius);
                font-size: 0.875rem;
            }
            
            input:focus {
                outline: none;
                border-color: var(--primary-color);
                box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
            }
            
            .empty-state {
                text-align: center;
                padding: 3rem;
                color: #6B7280;
                font-size: 0.875rem;
            }
        `;
    }

    connectedCallback() {
        this.render();
    }

    setData(products) {
        this.products = products;
        this.render();
    }

    render() {
        const style = `
            <style>
                ${EditableList.styles}
            </style>
        `;

        const html = `
            ${style}
            <div class="product-list">
                <form id="product-form">
                    <input type="text" name="name" placeholder="Nombre" required>
                    <input type="text" name="description" placeholder="Descripción" required>
                    <input type="number" name="price" placeholder="Precio" step="0.01" required>
                    <input type="number" name="stock" placeholder="Stock" required>
                    <button type="submit">Agregar Producto</button>
                </form>
                ${this.products.length === 0 ? `
                    <div class="empty-state">
                        No hay productos disponibles. Agrega uno nuevo.
                    </div>
                ` : `
                    <table>
                        <thead>
                            <tr>
                                <th>Nombre</th>
                                <th>Descripción</th>
                                <th>Precio</th>
                                <th>Stock</th>
                                <th>Acciones</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${this.products.map(product => `
                                <tr>
                                    <td>${product.name}</td>
                                    <td>${product.description}</td>
                                    <td>$${product.price}</td>
                                    <td>${product.stock}</td>
                                    <td class="actions">
                                        <button class="edit-btn" data-id="${product.id}">✏️</button>
                                        <button class="delete-btn" data-id="${product.id}">🗑️</button>
                                    </td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                `}
            </div>
        `;

        this.shadowRoot.innerHTML = html;
        this.addEventListeners();
    }

    addEventListeners() {
        const form = this.shadowRoot.getElementById('product-form');
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const product = {
                name: formData.get('name'),
                description: formData.get('description'),
                price: parseFloat(formData.get('price')),
                stock: parseInt(formData.get('stock'))
            };
            this.dispatchEvent(new CustomEvent('item-create', { detail: product }));
            form.reset();
        });

        this.shadowRoot.querySelectorAll('.edit-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const id = btn.dataset.id;
                const product = this.products.find(p => p.id === parseInt(id));
                this.dispatchEvent(new CustomEvent('item-edit', { detail: product }));
            });
        });

        this.shadowRoot.querySelectorAll('.delete-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const id = btn.dataset.id;
                this.dispatchEvent(new CustomEvent('item-delete', { detail: { id } }));
            });
        });
    }
}

customElements.define('editable-list', EditableList);