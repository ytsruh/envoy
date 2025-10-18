# Envoy

## Core Features:

- Variable Creation: Securely store environment variables using encryption.
- Project and Environment Management: Categorize environment variables into projects and environments (e.g., development, staging, production).
- Secure Sharing: Share environment variables securely with team members through access controls.
- AI Configuration Insights: AI-powered tool analyzes the created variables, cross-references them with current environment configurations and provides alerts of suspicious or incorrect values
- Format Export: Export environment variables in various formats (e.g., .env, JSON, YAML) for easy integration.
- Secure Access: User authentication with JWT-based password encryption.

## Style Guidelines:

- Primary color: Strong blue (#2979FF) for trust and security.
- Background color: Light blue (#E3F2FD), very desaturated, complementing the primary.
- Accent color: Purple (#7C4DFF), creating a vibrant contrast.
- Headline font: 'Space Grotesk', sans-serif, for a modern tech-oriented feel. Body font: 'Inter' sans-serif.
- Use lock and key icons to indicate the status of the environment variable with a color coded circle background.
- Clean and intuitive dashboard layout for easy variable management.
- Subtle animations for variable creation and saving confirmations.

## API Endpoints

### Authentication
- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login and receive JWT token
- Tokens expire in 7 days
- See [AUTH_API.md](docs/AUTH_API.md) for detailed documentation

### Other Endpoints
- `GET /health` - Health check endpoint
- `GET /hello` - Test greeting endpoint
- `POST /goodbye` - Test farewell endpoint

## Environment Variables

- `DB_PATH` - Path to SQLite database file
- `JWT_SECRET` - Secret key for JWT token signing (use a strong random string in production)
