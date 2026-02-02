// Utility Functions

// Format date
function formatDate(dateString) {
  const date = new Date(dateString);
  const options = { year: 'numeric', month: 'short', day: 'numeric' };
  return date.toLocaleDateString('en-US', options);
}

// Format cooking time
function formatCookingTime(minutes) {
  if (minutes < 60) {
    return `${minutes} min`;
  }
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
}

// Generate star rating HTML
function generateStars(rating, interactive = false) {
  const fullStars = Math.floor(rating);
  const hasHalf = rating % 1 >= 0.5;
  let html = '';

  for (let i = 1; i <= 5; i++) {
    if (interactive) {
      const activeClass = i <= rating ? 'active' : '';
      html += `<span class="rating-star ${activeClass}" data-rating="${i}">★</span>`;
    } else {
      if (i <= fullStars) {
        html += '★';
      } else if (i === fullStars + 1 && hasHalf) {
        html += '☆';
      } else {
        html += '☆';
      }
    }
  }

  return html;
}

// Get initials from name
function getInitials(name) {
  if (!name) return '?';
  return name
    .split(' ')
    .map(word => word[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

// Show notification
function showNotification(message, type = 'success') {
  const container = document.getElementById('notification-container') || createNotificationContainer();
  
  const notification = document.createElement('div');
  notification.className = `alert alert-${type}`;
  notification.innerHTML = `
    <span>${type === 'success' ? '[OK]' : type === 'error' ? '[X]' : '[i]'}</span>
    <span>${message}</span>
  `;

  container.appendChild(notification);

  setTimeout(() => {
    notification.style.opacity = '0';
    notification.style.transform = 'translateX(100%)';
    setTimeout(() => notification.remove(), 300);
  }, 4000);
}

function createNotificationContainer() {
  const container = document.createElement('div');
  container.id = 'notification-container';
  container.style.cssText = `
    position: fixed;
    top: 100px;
    right: 20px;
    z-index: 3000;
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-width: 400px;
  `;
  document.body.appendChild(container);
  return container;
}

// Debounce function
function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

// Parse URL params
function getUrlParams() {
  const params = new URLSearchParams(window.location.search);
  const result = {};
  for (const [key, value] of params) {
    result[key] = value;
  }
  return result;
}

// Validate email
function validateEmail(email) {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return re.test(email);
}

// Validate form
function validateForm(formData, rules) {
  const errors = {};

  for (const field in rules) {
    const value = formData[field];
    const fieldRules = rules[field];

    if (fieldRules.required && (!value || value.trim() === '')) {
      errors[field] = `${fieldRules.label || field} is required`;
      continue;
    }

    if (value && fieldRules.minLength && value.length < fieldRules.minLength) {
      errors[field] = `${fieldRules.label || field} must be at least ${fieldRules.minLength} characters`;
    }

    if (value && fieldRules.email && !validateEmail(value)) {
      errors[field] = 'Please enter a valid email address';
    }

    if (value && fieldRules.match && value !== formData[fieldRules.match]) {
      errors[field] = 'Passwords do not match';
    }
  }

  return errors;
}

// Show form errors
function showFormErrors(errors) {
  // Clear previous errors
  document.querySelectorAll('.form-error').forEach(el => el.remove());
  document.querySelectorAll('.form-input.error').forEach(el => el.classList.remove('error'));

  for (const field in errors) {
    const input = document.querySelector(`[name="${field}"]`);
    if (input) {
      input.classList.add('error');
      const errorEl = document.createElement('p');
      errorEl.className = 'form-error';
      errorEl.textContent = errors[field];
      input.parentNode.appendChild(errorEl);
    }
  }
}

// Loading state for buttons
function setButtonLoading(button, loading) {
  if (loading) {
    button.disabled = true;
    button.dataset.originalText = button.innerHTML;
    button.innerHTML = '<span class="spinner" style="width:20px;height:20px;border-width:2px;"></span>';
  } else {
    button.disabled = false;
    button.innerHTML = button.dataset.originalText;
  }
}

// Placeholder images
const PLACEHOLDER_IMAGES = {
  recipe: 'https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=800&auto=format&fit=crop&q=60',
  user: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200&auto=format&fit=crop&q=60'
};

function getRecipeImage(recipe) {
  if (recipe && recipe.image) return recipe.image;
  
  // Generate a consistent image based on recipe ID or title
  const images = [
    'https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=800&auto=format&fit=crop&q=60',
    'https://images.unsplash.com/photo-1567620905732-2d1ec7ab7445?w=800&auto=format&fit=crop&q=60',
    'https://images.unsplash.com/photo-1565299624946-b28f40a0ae38?w=800&auto=format&fit=crop&q=60',
    'https://images.unsplash.com/photo-1540189549336-e6e99c3679fe?w=800&auto=format&fit=crop&q=60',
    'https://images.unsplash.com/photo-1476224203421-9ac39bcb3327?w=800&auto=format&fit=crop&q=60',
    'https://images.unsplash.com/photo-1484723091951-59f36dee8b3d?w=800&auto=format&fit=crop&q=60'
  ];
  
  if (recipe && recipe.id) {
    const index = recipe.id.charCodeAt(recipe.id.length - 1) % images.length;
    return images[index];
  }
  
  return images[Math.floor(Math.random() * images.length)];
}

// Category icons mapping (text-based)
const CATEGORY_ICONS = {
  'breakfast': 'BF',
  'lunch': 'LN',
  'dinner': 'DN',
  'dessert': 'DS',
  'snack': 'SN',
  'drink': 'DR',
  'appetizer': 'AP',
  'soup': 'SP',
  'salad': 'SL',
  'pasta': 'PA',
  'meat': 'MT',
  'seafood': 'SF',
  'vegetarian': 'VG',
  'vegan': 'VN',
  'default': '**'
};

function getCategoryIcon(categoryName) {
  const name = (categoryName || '').toLowerCase();
  return CATEGORY_ICONS[name] || CATEGORY_ICONS.default;
}

// Export functions
window.formatDate = formatDate;
window.formatCookingTime = formatCookingTime;
window.generateStars = generateStars;
window.getInitials = getInitials;
window.showNotification = showNotification;
window.debounce = debounce;
window.getUrlParams = getUrlParams;
window.validateForm = validateForm;
window.showFormErrors = showFormErrors;
window.setButtonLoading = setButtonLoading;
window.getRecipeImage = getRecipeImage;
window.getCategoryIcon = getCategoryIcon;
