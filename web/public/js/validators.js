// TiendaSupported/web/public/js/validators.js

const validators = {
    required: (value) => {
        return value !== null && value !== undefined && value.toString().trim() !== '';
    },
    
    minLength: (value, min) => {
        return value.toString().length >= min;
    },
    
    number: (value) => {
        return !isNaN(parseFloat(value)) && isFinite(value);
    },
    
    positiveNumber: (value) => {
        return validators.number(value) && parseFloat(value) > 0;
    },
    
    integer: (value) => {
        return Number.isInteger(parseFloat(value));
    }
};

// Exportar la función para que pueda ser importada en app.js
export const validateProduct = (product) => {
    const errors = {};
    
    if (!validators.required(product.name)) {
        errors.name = 'El nombre es requerido';
    } else if (!validators.minLength(product.name, 3)) {
        errors.name = 'El nombre debe tener al menos 3 caracteres';
    }
    
    if (!validators.required(product.description)) {
        errors.description = 'La descripción es requerida';
    }
    
    if (!validators.positiveNumber(product.price)) {
        errors.price = 'El precio debe ser un número positivo';
    }
    
    if (!validators.integer(product.stock) || product.stock < 0) {
        errors.stock = 'El stock debe ser un número entero positivo';
    }
    
    return {
        isValid: Object.keys(errors).length === 0,
        errors
    };
};