import { serve } from '@hono/node-server'
import { Hono } from 'hono'
import { cors } from 'hono/cors'
import { logger } from 'hono/logger'
import { secureHeaders } from 'hono/secure-headers'
import { z } from 'zod'
import { zValidator } from '@hono/zod-validator'
import ical, { ICalAlarmType, ICalEventRepeatingFreq, ICalEventTransparency, ICalEventStatus } from 'ical-generator'
import { google, outlook, yahoo, office365, ics, type CalendarEvent as LinkEvent } from 'calendar-link'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc.js'
import timezone from 'dayjs/plugin/timezone.js'
import { v4 as uuidv4 } from 'uuid'
import { stripHtml } from 'string-strip-html'
import { rateLimiter } from 'hono-rate-limiter'

// --- Configuration ---
dayjs.extend(utc)
dayjs.extend(timezone)

const config = {
  port: Number(process.env.PORT) || 8001,
  corsOrigin: process.env.CORS_ORIGIN || '*',
  nodeEnv: process.env.NODE_ENV || 'development',
  rateLimit: {
    windowMs: 5 * 60 * 1000, // 15 minutes
    limit: 100
  }
}

const app = new Hono()

// --- Middleware ---
app.use('/*', cors({ origin: config.corsOrigin }))
app.use('/*', logger())
app.use('/*', secureHeaders())
app.use('/calendar/*', rateLimiter({
  windowMs: config.rateLimit.windowMs,
  limit: config.rateLimit.limit,
  standardHeaders: 'draft-6',
  keyGenerator: (c) => c.req.header('x-forwarded-for') || 'anonymous'
}))

// --- Validation Schemas ---

const eventSchema = z.object({
  // Core Data
  title: z.string().min(1, 'Title is required').max(200, 'Title too long'),
  start: z.string().datetime({ message: 'Start must be valid ISO 8601 (e.g. 2023-10-10T10:00:00Z)' }),
  end: z.string().datetime({ message: 'End must be valid ISO 8601' }),
  timezone: z.string().default('UTC'),
  
  // Content (with HTML sanitization)
  description: z.string().max(5000, 'Description too long').optional(),
  htmlDescription: z.string().max(10000, 'HTML description too long').optional().transform(val => {
    if (!val) return undefined
    const { result } = stripHtml(val, { 
      skipHtmlDecoding: false,
      stripTogetherWithTheirContents: ['script', 'style', 'iframe']
    })
    return result
  }),
  url: z.string().url().optional().or(z.literal('')).transform(val => val || undefined),
  location: z.string().max(500, 'Location too long').optional(),
  
  // Expert: Location & Geo
  geo: z.object({
    lat: z.number().min(-90).max(90),
    lon: z.number().min(-180).max(180)
  }).optional(),

  // Expert: Updates (UID & Sequence)
  uid: z.string().optional(),
  sequence: z.number().int().nonnegative().default(0), 
  
  // Expert: Status & Availability
  status: z.enum(['CONFIRMED', 'TENTATIVE', 'CANCELLED']).default('CONFIRMED'),
  busy: z.boolean().default(true),

  // Expert: Organizer
  organizer: z.object({
    name: z.string().min(1).max(100),
    email: z.string().email()
  }).optional(),

  // Expert: Attendees
  attendees: z.array(z.object({
    name: z.string().min(1).max(100),
    email: z.string().email(),
    rsvp: z.boolean().default(false),
    role: z.enum(['REQ-PARTICIPANT', 'OPT-PARTICIPANT', 'NON-PARTICIPANT']).default('REQ-PARTICIPANT'),
    status: z.enum(['ACCEPTED', 'DECLINED', 'TENTATIVE', 'NEEDS-ACTION']).default('NEEDS-ACTION')
  })).optional(),

  // Advanced Features
  allDay: z.boolean().optional(),
  categories: z.array(z.string().max(50)).max(10).optional(),
  
  alarms: z.array(z.object({
    minutes: z.number().int().positive().max(43200), // Max 30 days
    type: z.enum(['display', 'audio']).default('display'),
    description: z.string().max(200).optional()
  })).max(5).optional(),
  
  recurrence: z.object({
    freq: z.enum(['DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY']),
    interval: z.number().int().positive().max(1000).default(1),
    count: z.number().int().positive().max(999).optional(),
    until: z.string().datetime().optional(),
    byDay: z.array(z.enum(['MO', 'TU', 'WE', 'TH', 'FR', 'SA', 'SU'])).optional()
  }).optional(),

  options: z.object({
    method: z.enum(['PUBLISH', 'REQUEST', 'CANCEL']).default('PUBLISH')
  }).optional()
})
.refine((data) => dayjs(data.end).isAfter(dayjs(data.start)), {
  message: 'End date must be after start date',
  path: ['end']
})
.refine((data) => {
  if (data.recurrence?.until) {
    return dayjs(data.recurrence.until).isAfter(dayjs(data.start))
  }
  return true
}, {
  message: 'Recurrence end date must be after start date',
  path: ['recurrence', 'until']
})

const batchSchema = z.object({
  events: z.array(eventSchema).min(1, 'Must provide at least one event').max(100, 'Max 100 events per batch')
})

// Type inference
type CalendarEvent = z.infer<typeof eventSchema>
type BatchEvents = z.infer<typeof batchSchema>

// --- Helper Functions ---

const sanitizeHtml = (html?: string): string | undefined => {
  if (!html) return undefined
  const { result } = stripHtml(html, {
    skipHtmlDecoding: false,
    stripTogetherWithTheirContents: ['script', 'style', 'iframe', 'object', 'embed']
  })
  return result
}

const createIcalEvent = (calendar: any, data: CalendarEvent) => {
  const eventId = data.uid || uuidv4()
  
  const event = calendar.createEvent({
    start: dayjs(data.start).tz(data.timezone).toDate(),
    end: dayjs(data.end).tz(data.timezone).toDate(),
    timezone: data.timezone,
    summary: data.title,
    description: data.description,
    htmlDescription: data.htmlDescription,
    location: data.location,
    url: data.url,
    allDay: data.allDay,
    
    // Expert Fields
    id: eventId,
    sequence: data.sequence,
    status: ICalEventStatus[data.status],
    transparency: data.busy ? ICalEventTransparency.OPAQUE : ICalEventTransparency.TRANSPARENT,
    
    // Timestamps
    created: new Date(),
    lastModified: new Date(),
    
    // Categories
    categories: data.categories ? data.categories.map(c => ({ name: c })) : undefined
  })

  // Geo location
  if (data.geo) {
    event.geo({ lat: data.geo.lat, lon: data.geo.lon })
  }

  // Organizer
  if (data.organizer) {
    event.organizer({ 
      name: data.organizer.name, 
      email: data.organizer.email 
    })
  }

  // Attendees
  if (data.attendees) {
    data.attendees.forEach(attendee => {
      event.createAttendee({
        name: attendee.name,
        email: attendee.email,
        rsvp: attendee.rsvp,
        role: attendee.role as any,
        status: attendee.status as any
      })
    })
  }

  // Recurrence
  if (data.recurrence) {
    event.repeating({
      freq: ICalEventRepeatingFreq[data.recurrence.freq],
      interval: data.recurrence.interval,
      count: data.recurrence.count,
      until: data.recurrence.until ? dayjs(data.recurrence.until).toDate() : undefined,
      byDay: data.recurrence.byDay as any
    })
  }

  // Alarms
  if (data.alarms) {
    data.alarms.forEach(alarm => {
      event.createAlarm({
        type: alarm.type === 'audio' ? ICalAlarmType.audio : ICalAlarmType.display,
        trigger: alarm.minutes * 60,
        description: alarm.description
      })
    })
  }

  return eventId
}

const handleError = (error: unknown, context: string) => {
  console.error(`[${context}]`, error)
  
  const message = error instanceof Error ? error.message : 'Unknown error'
  const details = config.nodeEnv === 'development' ? { error: message } : undefined
  
  return {
    error: `Failed to ${context.toLowerCase()}`,
    ...details
  }
}

// --- Routes ---

app.get('/', (c) => c.html(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Calendar API - Documentation</title>
  <style>
    :root {
      --bg-color: #f8f9fa;
      --text-color: #212529;
      --accent-color: #667eea;
      --code-bg: #e9ecef;
      --card-bg: #ffffff;
      --border-color: #dee2e6;
      --shadow: 0 4px 6px rgba(0,0,0,0.1);
      --success: #28a745;
      --warning: #ffc107;
      --danger: #dc3545;
    }
    @media (prefers-color-scheme: dark) {
      :root {
        --bg-color: #0d1117;
        --text-color: #e6edf3;
        --accent-color: #8b9cf7;
        --code-bg: #161b22;
        --card-bg: #161b22;
        --border-color: #30363d;
        --shadow: 0 4px 6px rgba(0,0,0,0.4);
      }
    }
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      line-height: 1.6;
      color: var(--text-color);
      background: var(--bg-color);
    }
    .header {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      padding: 3rem 2rem;
      text-align: center;
    }
    .header h1 { font-size: 2.5rem; margin-bottom: 0.5rem; }
    .header p { font-size: 1.2rem; opacity: 0.9; }
    .badge {
      display: inline-block;
      padding: 4px 12px;
      border-radius: 20px;
      font-size: 0.85rem;
      font-weight: bold;
      margin: 0.5rem 0.25rem;
    }
    .badge.success { background: var(--success); color: white; }
    .container {
      max-width: 1200px;
      margin: 0 auto;
      padding: 2rem;
    }
    .section {
      background: var(--card-bg);
      padding: 2rem;
      border-radius: 8px;
      box-shadow: var(--shadow);
      margin-bottom: 2rem;
      border: 1px solid var(--border-color);
    }
    h2 {
      color: var(--accent-color);
      margin-bottom: 1rem;
      padding-bottom: 0.5rem;
      border-bottom: 2px solid var(--accent-color);
    }
    h3 { margin: 1.5rem 0 1rem; color: var(--text-color); }
    .endpoint {
      background: var(--code-bg);
      padding: 1rem;
      border-radius: 5px;
      font-family: monospace;
      font-size: 1rem;
      margin: 1rem 0;
      border: 1px solid var(--border-color);
      overflow-x: auto;
    }
    .method {
      color: #fff;
      padding: 4px 10px;
      border-radius: 4px;
      margin-right: 10px;
      font-weight: bold;
      font-size: 0.9rem;
    }
    .method.get { background: #28a745; }
    .method.post { background: #007bff; }
    code {
      background: var(--code-bg);
      padding: 3px 6px;
      border-radius: 4px;
      font-family: "SFMono-Regular", Consolas, monospace;
      font-size: 0.9em;
    }
    pre {
      background: var(--code-bg);
      padding: 1rem;
      border-radius: 5px;
      overflow-x: auto;
      border: 1px solid var(--border-color);
      margin: 1rem 0;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      margin: 1rem 0;
    }
    th, td {
      text-align: left;
      padding: 12px;
      border-bottom: 1px solid var(--border-color);
    }
    th {
      background: var(--code-bg);
      font-weight: 600;
    }
    .feature-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
      gap: 1rem;
      margin: 1.5rem 0;
    }
    .feature-card {
      background: var(--code-bg);
      padding: 1rem;
      border-radius: 8px;
      border: 1px solid var(--border-color);
    }
    .feature-card strong {
      color: var(--accent-color);
      display: block;
      margin-bottom: 0.5rem;
    }
    .footer {
      text-align: center;
      padding: 2rem;
      color: #6c757d;
      font-size: 0.9rem;
    }
    @media (max-width: 768px) {
      .header h1 { font-size: 2rem; }
      .header p { font-size: 1rem; }
      .container { padding: 1rem; }
      .section { padding: 1rem; }
      .feature-grid { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <div class="header">
    <h1>📅 Calendar API</h1>
    <p>Generate ICS files and calendar links for all major platforms</p>
    <div>
      <span class="badge success">v1.0.0</span>
      <span class="badge success">Production Ready</span>
    </div>
  </div>

  <div class="container">
    <div class="section">
      <h2>📖 Overview</h2>
      <p>A production-ready API for generating calendar events with support for ICS files, web calendar links, recurrence patterns, alarms, attendees, and more. Built with TypeScript, Hono, and Zod validation.</p>
    </div>

    <div class="section">
      <h2>🚀 Endpoints</h2>
      
      <h3>Health Check</h3>
      <div class="endpoint">
        <span class="method get">GET</span> /health
      </div>
      <p>Returns API health status and uptime information.</p>

      <h3>Validate Event</h3>
      <div class="endpoint">
        <span class="method post">POST</span> /calendar/validate
      </div>
      <p>Validate event data without generating files. Useful for testing schemas.</p>

      <h3>Generate Calendar Links</h3>
      <div class="endpoint">
        <span class="method post">POST</span> /calendar/links
      </div>
      <p>Generate add-to-calendar links for Google, Outlook, Yahoo, Office365, and ICS.</p>

      <h3>Generate ICS File</h3>
      <div class="endpoint">
        <span class="method post">POST</span> /calendar/ics
      </div>
      <p>Generate a single .ics file for download with full event details.</p>

      <h3>Batch ICS Generation</h3>
      <div class="endpoint">
        <span class="method post">POST</span> /calendar/batch
      </div>
      <p>Generate an .ics file containing multiple events (up to 100).</p>

      <h3>All-in-One</h3>
      <div class="endpoint">
        <span class="method post">POST</span> /calendar/all
      </div>
      <p>Get both ICS content and provider links in a single response.</p>
    </div>

    <div class="section">
      <h2>⚡ Features</h2>
      <div class="feature-grid">
        <div class="feature-card">
          <strong>🔁 Recurrence</strong>
          Support for DAILY, WEEKLY, MONTHLY, YEARLY patterns with intervals and end dates
        </div>
        <div class="feature-card">
          <strong>⏰ Multiple Alarms</strong>
          Up to 5 alarms per event with custom descriptions
        </div>
        <div class="feature-card">
          <strong>👥 Attendee Tracking</strong>
          RSVP support with participant roles and status
        </div>
        <div class="feature-card">
          <strong>🌍 Timezone Support</strong>
          Per-event timezone configuration
        </div>
        <div class="feature-card">
          <strong>📍 Geo-location</strong>
          Latitude/longitude coordinates for events
        </div>
        <div class="feature-card">
          <strong>🏷️ Categories/Tags</strong>
          Organize events with custom categories
        </div>
        <div class="feature-card">
          <strong>🛡️ HTML Sanitization</strong>
          Automatic stripping of dangerous HTML tags
        </div>
        <div class="feature-card">
          <strong>🚦 Rate Limiting</strong>
          ${config.rateLimit.limit} requests per 5 minutes per IP
        </div>
      </div>
    </div>

    <div class="section">
      <h2>📝 Example Request</h2>
      <h3>Simple Event</h3>
      <pre><code>POST /calendar/ics
Content-Type: application/json

{
  "title": "Team Meeting",
  "description": "Quarterly planning session",
  "start": "2026-02-15T10:00:00Z",
  "end": "2026-02-15T11:00:00Z",
  "location": "Conference Room A",
  "timezone": "America/New_York"
}</code></pre>

      <h3>Advanced Event with Recurrence</h3>
      <pre><code>POST /calendar/ics
Content-Type: application/json

{
  "title": "Weekly Standup",
  "start": "2026-02-15T09:00:00Z",
  "end": "2026-02-15T09:30:00Z",
  "timezone": "UTC",
  "alarms": [
    { "minutes": 15, "type": "display", "description": "Meeting in 15 min" }
  ],
  "recurrence": {
    "freq": "WEEKLY",
    "interval": 1,
    "byDay": ["MO", "WE", "FR"],
    "count": 20
  },
  "categories": ["work", "standup"],
  "organizer": {
    "name": "John Doe",
    "email": "john@example.com"
  },
  "attendees": [
    {
      "name": "Jane Smith",
      "email": "jane@example.com",
      "rsvp": true,
      "role": "REQ-PARTICIPANT",
      "status": "NEEDS-ACTION"
    }
  ]
}</code></pre>

      <h3>Response (ICS File)</h3>
      <pre><code>Content-Type: text/calendar; charset=utf-8
Content-Disposition: attachment; filename="event-{uuid}.ics"
X-Event-ID: {uuid}

BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Calendar API//EN
BEGIN:VEVENT
...</code></pre>
    </div>

    <div class="section">
      <h2>📋 Event Schema</h2>
      <table>
        <thead>
          <tr>
            <th>Field</th>
            <th>Type</th>
            <th>Required</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td><code>title</code></td>
            <td>string</td>
            <td>✅</td>
            <td>Event title (max 200 chars)</td>
          </tr>
          <tr>
            <td><code>start</code></td>
            <td>ISO 8601</td>
            <td>✅</td>
            <td>Start date/time</td>
          </tr>
          <tr>
            <td><code>end</code></td>
            <td>ISO 8601</td>
            <td>✅</td>
            <td>End date/time</td>
          </tr>
          <tr>
            <td><code>description</code></td>
            <td>string</td>
            <td>❌</td>
            <td>Plain text description (max 5000 chars)</td>
          </tr>
          <tr>
            <td><code>location</code></td>
            <td>string</td>
            <td>❌</td>
            <td>Event location</td>
          </tr>
          <tr>
            <td><code>timezone</code></td>
            <td>string</td>
            <td>❌</td>
            <td>IANA timezone (default: UTC)</td>
          </tr>
          <tr>
            <td><code>allDay</code></td>
            <td>boolean</td>
            <td>❌</td>
            <td>All-day event flag</td>
          </tr>
          <tr>
            <td><code>url</code></td>
            <td>URL</td>
            <td>❌</td>
            <td>Associated URL</td>
          </tr>
          <tr>
            <td><code>status</code></td>
            <td>enum</td>
            <td>❌</td>
            <td>CONFIRMED | TENTATIVE | CANCELLED</td>
          </tr>
          <tr>
            <td><code>categories</code></td>
            <td>string[]</td>
            <td>❌</td>
            <td>Event tags (max 10)</td>
          </tr>
          <tr>
            <td><code>alarms</code></td>
            <td>object[]</td>
            <td>❌</td>
            <td>Reminder alarms (max 5)</td>
          </tr>
          <tr>
            <td><code>recurrence</code></td>
            <td>object</td>
            <td>❌</td>
            <td>Recurrence pattern</td>
          </tr>
          <tr>
            <td><code>organizer</code></td>
            <td>object</td>
            <td>❌</td>
            <td>Event organizer details</td>
          </tr>
          <tr>
            <td><code>attendees</code></td>
            <td>object[]</td>
            <td>❌</td>
            <td>Event attendees</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="section">
      <h2>🔗 Calendar Links Response</h2>
      <pre><code>POST /calendar/links

Response:
{
  "success": true,
  "eventId": "550e8400-e29b-41d4-a716-446655440000",
  "links": {
    "google": "https://calendar.google.com/calendar/render?action=TEMPLATE&...",
    "outlook": "https://outlook.live.com/calendar/0/deeplink/compose?path=/calendar/action/compose&...",
    "yahoo": "https://calendar.yahoo.com/?v=60&...",
    "office365": "https://outlook.office.com/calendar/0/deeplink/compose?...",
    "ics": "data:text/calendar;charset=utf8,BEGIN:VCALENDAR..."
  }
}</code></pre>
    </div>

    <div class="section">
      <h2>🛡️ Security & Rate Limits</h2>
      <ul style="list-style: none; padding: 0;">
        <li style="padding: 0.5rem 0;">✅ HTML sanitization with string-strip-html</li>
        <li style="padding: 0.5rem 0;">✅ Zod schema validation on all inputs</li>
        <li style="padding: 0.5rem 0;">✅ Secure headers middleware</li>
        <li style="padding: 0.5rem 0;">✅ CORS protection</li>
        <li style="padding: 0.5rem 0;">✅ Rate limiting: ${config.rateLimit.limit} req / 5 min</li>
        <li style="padding: 0.5rem 0;">✅ Request logging</li>
      </ul>
    </div>

    <div class="section">
      <h2>💻 JavaScript Integration</h2>
      <pre><code>// Generate ICS file
async function createCalendarEvent() {
  const response = await fetch('/calendar/ics', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title: 'My Event',
      start: '2026-03-01T14:00:00Z',
      end: '2026-03-01T15:00:00Z',
      timezone: 'America/New_York'
    })
  });
  
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'event.ics';
  a.click();
}

// Get calendar links
async function getCalendarLinks() {
  const response = await fetch('/calendar/links', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title: 'My Event',
      start: '2026-03-01T14:00:00Z',
      end: '2026-03-01T15:00:00Z'
    })
  });
  
  const data = await response.json();
  console.log(data.links.google); // Open in new tab
}</code></pre>
    </div>

    <div class="section">
      <h2>📊 Status Codes</h2>
      <table>
        <thead>
          <tr>
            <th>Code</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td><code>200 OK</code></td>
            <td>Request successful</td>
          </tr>
          <tr>
            <td><code>400 Bad Request</code></td>
            <td>Validation failed - check error details</td>
          </tr>
          <tr>
            <td><code>404 Not Found</code></td>
            <td>Endpoint doesn't exist</td>
          </tr>
          <tr>
            <td><code>429 Too Many Requests</code></td>
            <td>Rate limit exceeded</td>
          </tr>
          <tr>
            <td><code>500 Internal Server Error</code></td>
            <td>Server error occurred</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <div class="footer">
    <p>Calendar API v1.0.0 • Built with Hono, TypeScript & Zod • Production Ready</p>
  </div>
</body>
</html>
`))

app.get('/health', (c) => c.json({ 
  status: 'healthy', 
  timestamp: new Date().toISOString(),
  uptime: process.uptime(),
  version: '3.0.0'
}))

// Validate endpoint
app.post('/calendar/validate', zValidator('json', eventSchema), (c) => {
  const data = c.req.valid('json')
  return c.json({
    valid: true,
    message: 'Event data is valid',
    eventId: data.uid || 'will-be-generated',
    parsedData: {
      title: data.title,
      start: dayjs(data.start).format(),
      end: dayjs(data.end).format(),
      timezone: data.timezone,
      hasRecurrence: !!data.recurrence,
      hasAlarms: !!data.alarms?.length,
      hasAttendees: !!data.attendees?.length
    }
  })
})

// Web Links Handler
app.post('/calendar/links', zValidator('json', eventSchema), async (c) => {
  const data = c.req.valid('json')
  
  try {
    const linkEvent: LinkEvent = {
      title: data.title,
      description: data.description || sanitizeHtml(data.htmlDescription) || '',
      start: data.start,
      end: data.end,
      location: data.location,
      url: data.url,
      allDay: data.allDay
    }

    return c.json({
      success: true,
      eventId: data.uid || uuidv4(),
      links: {
        google: google(linkEvent),
        outlook: outlook(linkEvent),
        yahoo: yahoo(linkEvent),
        office365: office365(linkEvent),
        ics: ics(linkEvent)
      }
    })
  } catch (error) {
    return c.json(handleError(error, 'Generate calendar links'), 500)
  }
})

// Single ICS Handler
app.post('/calendar/ics', zValidator('json', eventSchema), async (c) => {
  const data = c.req.valid('json')
  
  try {
    const calendar = ical({
      name: data.title,
      method: data.options?.method as any,
      timezone: data.timezone
    })

    const eventId = createIcalEvent(calendar, data)
    const icsContent = calendar.toString()

    return c.text(icsContent, 200, {
      'Content-Type': 'text/calendar; charset=utf-8',
      'Content-Disposition': `attachment; filename="event-${eventId}.ics"`,
      'X-Event-ID': eventId
    })
  } catch (error) {
    return c.json(handleError(error, 'Generate ICS file'), 500)
  }
})

// Batch ICS Handler
app.post('/calendar/batch', zValidator('json', batchSchema), async (c) => {
  const { events } = c.req.valid('json')
  
  try {
    const calendar = ical({ 
      name: 'Batch Events',
      timezone: events[0]?.timezone || 'UTC'
    })

    const eventIds: string[] = []
    events.forEach(eventData => {
      const eventId = createIcalEvent(calendar, eventData)
      eventIds.push(eventId)
    })

    const icsContent = calendar.toString()
    const timestamp = dayjs().unix()

    return c.text(icsContent, 200, {
      'Content-Type': 'text/calendar; charset=utf-8',
      'Content-Disposition': `attachment; filename="batch-${timestamp}.ics"`,
      'X-Event-Count': events.length.toString(),
      'X-Event-IDs': eventIds.join(',')
    })
  } catch (error) {
    return c.json(handleError(error, 'Generate batch ICS file'), 500)
  }
})

// Combined Handler (ICS + Links)
app.post('/calendar/all', zValidator('json', eventSchema), async (c) => {
  const data = c.req.valid('json')
  
  try {
    // Generate ICS
    const calendar = ical({
      name: data.title,
      method: data.options?.method as any,
      timezone: data.timezone
    })
    const eventId = createIcalEvent(calendar, data)
    const icsContent = calendar.toString()

    // Generate Links
    const linkEvent: LinkEvent = {
      title: data.title,
      description: data.description || sanitizeHtml(data.htmlDescription) || '',
      start: data.start,
      end: data.end,
      location: data.location,
      url: data.url,
      allDay: data.allDay
    }

    return c.json({
      success: true,
      eventId,
      ics: {
        content: icsContent,
        filename: `event-${eventId}.ics`
      },
      links: {
        google: google(linkEvent),
        outlook: outlook(linkEvent),
        yahoo: yahoo(linkEvent),
        office365: office365(linkEvent),
        ics: ics(linkEvent)
      }
    })
  } catch (error) {
    return c.json(handleError(error, 'Generate calendar data'), 500)
  }
})

// 404 Handler
app.notFound((c) => c.json({ error: 'Endpoint not found' }, 404))

// Error Handler
app.onError((err, c) => {
  console.error('[Global Error]', err)
  return c.json({
    error: 'Internal server error',
    ...(config.nodeEnv === 'development' && { details: err.message })
  }, 500)
})

// --- Server Startup ---
console.log(`🚀 Calendar API v1.0.0`)
console.log(`📍 Running on http://localhost:${config.port}`)
console.log(`🌍 Environment: ${config.nodeEnv}`)
console.log(`🛡️  Rate Limit: ${config.rateLimit.limit} requests per 5 minutes`)

serve({ fetch: app.fetch, port: config.port })