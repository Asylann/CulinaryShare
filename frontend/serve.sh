#!/bin/bash
# Start the frontend development server

echo "🍳 CulinaryShare Frontend"
echo "========================="
echo ""
echo "Starting development server..."
echo ""

cd "$(dirname "$0")"

# Check if Python is available
if command -v python3 &> /dev/null; then
    echo "📂 Serving frontend at: http://localhost:3000"
    echo "🔗 Open in your browser: http://localhost:3000"
    echo ""
    echo "Press Ctrl+C to stop the server"
    echo ""
    python3 -m http.server 3000
elif command -v python &> /dev/null; then
    echo "📂 Serving frontend at: http://localhost:3000"
    echo "🔗 Open in your browser: http://localhost:3000"
    echo ""
    echo "Press Ctrl+C to stop the server"
    echo ""
    python -m http.server 3000
else
    echo "❌ Python is not installed. Please install Python or use another HTTP server."
    echo ""
    echo "Alternatives:"
    echo "  - npx serve ."
    echo "  - php -S localhost:3000"
    echo "  - Open index.html directly in your browser"
    exit 1
fi
