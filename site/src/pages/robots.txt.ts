import { absoluteUrl } from '../lib/site';

const base = import.meta.env.BASE_URL;

export function GET() {
  return new Response(
    `User-agent: *
Allow: ${base}

Sitemap: ${absoluteUrl('sitemap.xml')}
`,
    {
      headers: {
        'Content-Type': 'text/plain',
      },
    },
  );
}
