export type Variable = {
  id: string;
  key: string;
  value: string;
  status: 'secure' | 'warning' | 'insecure';
  comment: string | null;
};

export type Environment = {
  id: 'development' | 'staging' | 'production';
  name: 'Development' | 'Staging' | 'Production';
  variables: Variable[];
};

export type Project = {
  id: string;
  name: string;
  environments: Environment[];
};

export const projects: Project[] = [
  {
    id: 'proj-1',
    name: 'WebApp_Frontend',
    environments: [
      {
        id: 'development',
        name: 'Development',
        variables: [
          { id: 'var-1', key: 'NEXT_PUBLIC_API_URL', value: 'http://localhost:3001/api', status: 'warning', comment: 'Local dev endpoint' },
          { id: 'var-2', key: 'DATABASE_URL', value: 'postgresql://user:pass@localhost:5432/db', status: 'insecure', comment: 'Should use a secret manager' },
          { id: 'var-3', key: 'SESSION_SECRET', value: 'a-very-long-and-random-string-for-dev', status: 'secure', comment: null },
        ],
      },
      {
        id: 'staging',
        name: 'Staging',
        variables: [
          { id: 'var-4', key: 'NEXT_PUBLIC_API_URL', value: 'https://staging.api.envizo.dev/api', status: 'secure', comment: 'Staging endpoint' },
          { id: 'var-5', key: 'DATABASE_URL', value: 'postgresql://user:***@staging-db.aws.com:5432/db', status: 'secure', comment: 'Uses secrets manager' },
          { id: 'var-6', key: 'SESSION_SECRET', value: 'a-very-long-and-random-string-for-staging', status: 'secure', comment: null },
        ],
      },
      {
        id: 'production',
        name: 'Production',
        variables: [
           { id: 'var-7', key: 'NEXT_PUBLIC_API_URL', value: 'https://api.envizo.dev/api', status: 'secure', comment: 'Production endpoint' },
          { id: 'var-8', key: 'DATABASE_URL', value: 'postgresql://user:***@prod-db.aws.com:5432/db', status: 'secure', comment: 'Uses secrets manager' },
          { id: 'var-9', key: 'SESSION_SECRET', value: 'a-very-long-and-random-string-for-production', status: 'secure', comment: null },
        ],
      },
    ],
  },
  {
    id: 'proj-2',
    name: 'Mobile_Backend',
    environments: [
      { id: 'development', name: 'Development', variables: [] },
      { id: 'staging', name: 'Staging', variables: [] },
      { id: 'production', name: 'Production', variables: [] },
    ],
  },
];
