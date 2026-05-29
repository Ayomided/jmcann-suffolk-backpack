#let fonts = (
  captions: "Anuphan",
  default: "Anuphan",
  serif: "Rockwell",
  heading: "Insignia LT Std",
)

#let colors = (
  brand: rgb("#000"),
)

#let tag(color: colors.brand.lighten(80%), tag_token) = [
  #set text(size: 9pt, font: fonts.captions, weight: "medium", fill: white)
  #rect(radius: 5pt, fill: color)[#tag_token]
]

#let process_step(
  number,
  title,
  items,
  color: red.lighten(20%),
  text_color: white,
  accent: rgb("#e0157a"),
) = {
  box(width: 120pt)[
    #stack(
      dir: ttb,
      spacing: 2pt,
      rect(
        width: 120pt,
        height: 120pt,
        radius: (top-left: 60pt, top-right: 60pt, bottom-left: 0pt, bottom-right: 0pt),
        fill: color,
      )[
        #place(left, dy: 30pt, dx: -12pt)[
          #stack(dir: ltr, curve(
            fill: white,
            stroke: none,
            curve.move((0pt, 13pt)),
            curve.line((20pt, 13pt)),
            curve.line((14pt, 6pt)),
            curve.line((20pt, 6pt)),
            curve.line((28pt, 16pt)),
            curve.line((20pt, 26pt)),
            curve.line((14pt, 26pt)),
            curve.line((20pt, 19pt)),
            curve.line((0pt, 19pt)),
            curve.close(),
          ))
        ]
        #align(center + horizon)[
          #set text(fill: text_color, size: 52pt, weight: "extrabold", font: fonts.heading)
          #number
        ]
      ],
      align(left)[
        #rect(
          width: 120pt,
          height: 100pt,
          fill: color,
          inset: (left: 4pt, right: 2pt),
        )[
          #set text(fill: white, size: 8pt)
          #rect(fill: white)[
            #set text(fill: color, size: 10pt)
            #text(weight: "bold", size: 12pt)[#title]
          ]
          #v(6pt)
          #for item in items [
            #text[- #item]
          ]
        ]
      ],
    )
  ]
}

#let process_flow(steps, color: rgb("#1a2e2a"), wrap: 4) = {
  align(center)[
    #stack(
      dir: ltr,
      spacing: 40pt / wrap,
      ..steps
        .enumerate()
        .map(((i, s)) => {
          let pill = process_step(
            str(i + 1),
            s.at(0),
            s.at(1),
            color: color,
          )
          (pill,)
        })
        .flatten(),
    )
  ]
}

#let make_clip(scale_pct: 45%) = {
  let s = scale_pct / 100%

  // base dimensions (at 100%)
  let base_w = 70pt
  let base_h = 175pt
  let box_w = base_w * s
  let box_h = base_h * s

  let cs = stroke(
    paint: rgb("#e0157a"),
    thickness: 18pt * s,
    cap: "round",
    join: "round",
  )

  let back = box(width: box_w, height: box_h)[
    #scale(scale_pct, origin: top + left)[
      #box(width: base_w, height: base_h)[
        #place(top + left)[
          #curve(
            stroke: cs,
            curve.move((15pt, 115pt)),
            curve.line((15pt, 35pt)),
            curve.cubic((15pt, 8pt), (55pt, 8pt), (55pt, 35pt)),
            curve.line((55pt, 115pt)),
          )
        ]
      ]
    ]
  ]

  let front = box(width: box_w, height: box_h)[
    #scale(scale_pct, origin: top + left)[
      #box(width: base_w, height: base_h)[
        #place(top + left)[
          #curve(
            stroke: cs,
            curve.move((28pt, 65pt)),
            curve.line((28pt, 120pt)),
            curve.cubic((28pt, 140pt), (55pt, 140pt), (55pt, 120pt)),
            curve.line((55pt, 85pt)),
          )
        ]
        #place(top + left, dx: 5pt, dy: -8pt)[
          #line(
            start: (63pt, 85pt),
            end: (63pt, 120pt),
            stroke: cs,
          )
        ]
      ]
    ]
  ]

  (back, front, box_w, box_h)
}

#let clipped_card(content, scale_pct: 45%, card_width: 180pt, card_color: white, rot: -3deg) = {
  let (back, front, clip_w, clip_h) = make_clip(scale_pct: scale_pct)
  let peek = clip_h * 0.65

  let card = rect(
    width: card_width,
    fill: card_color,
    inset: 14pt,
    radius: 2pt,
    stroke: none,
  )[#content]

  context {
    let card_h = measure(card).height

    block(
      width: card_width,
      height: card_h + peek,
      clip: false,
    )[
      #place(top + left, dx: card_width / 2 - clip_w / 2, dy: 0pt)[#back]
      #place(top + left, dy: peek)[#rotate(rot, origin: top + center)[#card]]
      #place(top + left, dx: card_width / 2 - clip_w / 2, dy: 0pt)[#front]
    ]
  }
}

#let step_colors = (
  staff: rgb("#8B0000"),
  system: rgb("#636463"),
  manager: rgb("#007B8A"),
  end: rgb("#2d3142"),
)

#let step_size = 40pt
#let arrow_color = rgb("#636463")
#let arrow_stroke = stroke(paint: arrow_color, thickness: 1.5pt)

// Draw the shape for a given action type
#let draw_shape(kind) = {
  let c = step_colors.at(kind)
  if kind == "staff" [
    #circle(radius: step_size / 2, fill: c, stroke: none)
  ] else if kind == "end" [
    #polygon.regular(vertices: 3, size: step_size, fill: c, stroke: none)
  ] else [
    #rect(width: step_size, height: step_size, fill: c, stroke: none, radius: 4pt)
  ]
}

// Arrow between steps
#let arrow = context [
  #line(length: 30pt, stroke: arrow_stroke)
  #place(right + horizon, dx: 2pt)[
    #polygon(
      fill: arrow_color,
      (0pt, 0pt),
      (-8pt, -4pt),
      (-8pt, 4pt),
    )
  ]
]

// A single step: shape + label below
#let step(kind, label, icon: none) = [
  #stack(
    dir: ttb,
    spacing: 6pt,
    box(width: step_size, height: step_size)[
      #place(center + horizon)[#draw_shape(kind)]
      #if icon != none [
        #place(center + horizon)[
          #set text(fill: white, size: 10pt)
          #icon
        ]
      ]
    ],
    box(width: 60pt)[
      #set text(size: 7pt, weight: "regular")
      #align(center)[#label]
    ],
  )
]

#let conf(title, doc) = {
  set text(font: fonts.default, lang: "en", weight: "regular")

  let page_margin = 48pt

  set page(margin: page_margin)
  set page(numbering: "1")
  set page(
    header: [
      #set text(font: fonts.serif, size: 6pt, weight: "light")
      David Adediji | KTP Associate Candidate #h(1fr) #title
    ],
    footer: context [
      #let (num,) = counter(page).get()
      #pad(x: -page_margin)[
        #rect(
          width: 100%,
          height: 100%,
          inset: (x: page_margin, y: 4pt),
          fill: colors.brand,
        )[
          #set text(font: fonts.serif, fill: white, size: 10pt)
          #if calc.rem-euclid(num, 2) == 0 [
            #align(horizon + right)[Page #num]
          ] else [
            #align(horizon + left)[Page #num]
          ]
        ]
      ]
    ],
  )

  show emph: it => {
    set text(style: "italic")
    stack(dir: ttb, spacing: 20pt, rect(
      stroke: (left: 2pt + yellow),
      fill: yellow.lighten(80%),
    )[> #it.body])
  }

  show figure.caption: it => [
    #set text(size: 9pt, font: fonts.captions, weight: "thin")
    #box(baseline: 30%)[#tag(it.supplement)] #it.body
  ]

  [
    #v(20%)
    #text(size: 22pt, weight: "semibold")[Backpack App Requirements Document]
  ]
  pagebreak()

  set heading(numbering: "1.")
  outline()

  pagebreak()

  show heading: set text(font: fonts.default, size: 12pt, weight: "medium")
  show heading: it => block[
    #stack(
      dir: ttb,
      spacing: 8pt,
      [#counter(heading).display(it.numbering) #upper(it.body)],
      line(
        length: 100%,
        stroke: 1pt + black,
      ),
    )
  ]

  show heading.where(level: 2): it => block[
    #stack(
      dir: ttb,
      spacing: 8pt,
      [#counter(heading).display(it.numbering) #upper(it.body)],
      line(
        length: 100%,
        stroke: 1pt + black,
      ),
    )
  ]

  set table(stroke: (bottom: 0.5pt + gray, left: none, right: none, top: none), inset: 1em, fill: (x, y) => {
    if calc.rem-euclid(y, 2) == 0 { gray.lighten(80%) }
  })
  show table.header: set text(weight: "bold")

  doc
}

#show: doc => conf("UoS | J McCann Backpack App - Requirements Document", doc)


= Introduction
McCann wishes to better understand the cost of a completed job, the team believes an application enabling operatives to fill in job details would assist with this, as well as enabling the intended job cost calculation to be performed by the QS

This document presents the requirements of the Backpack application specification that will be satisfy, from a technical and system operations perspective in relation to the business process it will support.

The Backpack app is developed by the KTP Associate in collaboration with the Business Partner Supervisor at McCann and the Knowledge Base Supervisor, University of Suffolk.

== Scope
This requirements specification applies to the task version completed as part of the KTP recruitment process, it is built to demonstrate my existing experience in Software engineering.

= Roles & Responsibilities
#table(
  columns: (auto, 1fr),
  [*Role*], [*Responsible for*],

  [Operative],
  [
    - Record and update site visit record
    - Record materials used during site visits
  ],

  [QS],
  [
    - Review and approve visit records
    - Manage rates for operatives and materials
  ],
)

= Definitions
Brief business and technical definitions specific to the business process that may be used in this document in relation to the system are given bekow:

#table(
  columns: (auto, 1fr),
  [*Operative*],
  [Worker on site who completes hands on work within a location, e.g. a "job" such as repairing a streetlight],

  [*Quantity Surveyor (QS)*], [An Office-bases staff who completes the financial parts of each job],

  [*Site visit*], [A session of time where an operative completes hands on work within a location],
)

= Overview of the System
== Intended use of the system
The Backpack app will be used by Operatives and QSs to report and manage calculations of completed jobs respectively.

Records of site visits will be recorded by the Operative, including any resources used, the cost of these items are calculated as well as cost of working operatives on the job, this is presented for the QS to complete the financial part of the job. The system will also allow for the QSs to create jobs with planned costs to compare against actuals.

The key purpose of the system is for McCann and her teams to understand better the cost of completed jobs.

== Description of Busines Process
#figure(image("docs/img/operative_process.png", width: 25%), caption: "Site visit recording")

#figure(image("docs/img/qs_process.png", width: 45%), caption: "QS cost management")


#let req-counters = (
  "AUTH": counter("req-AUTH"),
  "JOB": counter("req-JOB"),
  "SESSION": counter("req-SESSION"),
  "RESOURCE": counter("req-RESOURCE"),
  "COST": counter("req-COST"),
)

#let requirement-id(reqtype) = context {
  req-counters.at(reqtype).step()
  let n = req-counters.at(reqtype).get().first() + 1
  [R-MCANN-#reqtype\-#n]
}

= Requirements
#table(
  columns: (auto, 1fr),
  [*Reference*], [*Requirement*],
  [#requirement-id("AUTH")], [Users must authenticate with an email address and password],
  [#requirement-id("AUTH")],
  [Role-based access control is enforced at the middleware level; such that Operatives can't access QS routes and vice versa],

  [#requirement-id("JOB")],
  [QS users can create jobs with a reference number, name, site, start date and expected headcount],

  [#requirement-id("JOB")], [Jobs have a status lifecycle: Draft → Active → Completed → Approved],
  [#requirement-id("JOB")], [QS users can update job status and mark jobs as complete],
  [#requirement-id("JOB")], [Each job is associated with a site and tracks its sessions, operatives and total cost],

  [#requirement-id("SESSION")],
  [Operatives create a session to record a site visit, capturing job, start time and optional notes],

  [#requirement-id("SESSION")],
  [An operative may only have one open (unsubmitted) session at a time, this is enforced at both application and database level, reflecting that an operative can only be physically present at one site simultaneously],

  [#requirement-id("SESSION")],
  [A session is submitted by the operative when the site visit is complete for the day, recording the end time and submitting operative],

  [#requirement-id("SESSION")], [Submitted sessions cannot be modified],

  [#requirement-id("RESOURCE")],
  [Operatives log resources used within a session across three categories; Materials, Tools, and Mechanical],

  [#requirement-id("RESOURCE")],
  [The rate applicable to each resource is snapshotted at the time of recording to preserve historical accuracy regardless of future rate changes],

  [#requirement-id("RESOURCE")], [Calculated cost is stored alongside the rate snapshot],

  [#requirement-id("COST")],
  [The system automatically calculates costs across all sessions, operatives and resources when a job is marked complete],

  [#requirement-id("COST")],
  [Operative costs are derived from the snapshotted rate and the duration between arrival and departure times],
)

= Implementation
== Language
The Backpack app is implemented in Go, a statically typed, compiled language developed by Google. Go was selected for its simplicity, strong standard library, and performance characteristics that make it well suited for web applications. The application compiles to a single self-contained binary which includes the HTTP server, route handlers, template rendering engine and database access layer. No external runtime or interpreter is required to run the application on the server.

The standard library `net/http` package is used to implement the HTTP server and routing, and `html/template` is used for server-side HTML rendering. This approach was chosen over a JavaScript framework to reduce complexity, eliminate a frontend build step, and ensure the application remains performant on low-powered mobile devices — a consideration for field operatives working in environments with poor connectivity.

The database layer uses SQLite via the `mattn/go-sqlite3` driver, abstracted behind an interface such that the underlying database can be replaced with PostgreSQL without changes to the application or handler code.

== Screens
The following screens were implemented and are included as part of this submission:

- *Login* — email and password authentication with validation error states, available to both roles
- *QS Jobs List* — tabular view of active and completed jobs, desktop optimised, accessible to QS users
- *Operative Jobs List* — card-based view of active and completed jobs, mobile optimised, accessible to operative users
- *Job Detail* — role-aware view showing job details, session history and operatives; QS users see cost summary and status actions, operatives see the Mark Session action
- *Create Job* — form for QS users to create a new job with reference, name, site, start date and headcount, including field-level validation error states
- *Create Session* — form for operatives to record a site visit against a job with start time and optional notes
- *Session Detail* — role-aware view showing session details with materials, tools and mechanical resources logged; QS users see calculated costs per line item and session total

== Testing
The operative session logging screen was tested using Rod, a Go browser automation library built on the Chrome DevTools Protocol. Rod was selected as it integrates directly with the Go test toolchain without requiring a separate WebDriver process.

The following test cases were implemented:

- *Login renders correctly* — asserts that the email input, password input and submit button are present in the DOM
- *Login with valid credentials* — submits valid credentials and asserts the browser is redirected to the dashboard
- *Login with invalid credentials* — submits an incorrect password and asserts the browser remains on the login page and a validation error is displayed
- *Login with empty fields* — submits an empty form and asserts field-level error elements are rendered for both email and password

The Money type, which handles all monetary arithmetic in the application, is covered by unit tests asserting correct behaviour of addition, multiplication and serialisation in minor currency units.

== Deployment and Repository
The source code is hosted on Codeberg, a source-available Git hosting platform, at the following address:

`https://codeberg.org/ayomided/jmcann-suffolk-backpack`

A live deployment of the application is available at:

`https://backpack.adediiji.uk`

The application is deployed to a Hetzner VPS running Ubuntu. The Go binary and static assets are built and transferred to the server over SSH as part of the CI/CD pipeline, which is triggered on every push to the main branch. The pipeline performs the following steps in order:

+ Run lints — code style and quality checks using `golangci-lint`
+ Run tests — unit and UI tests using the Go test toolchain and Rod
+ Build binary — `go build` compiles the application to a single binary
+ Deploy — the binary and static assets are copied to the server over SSH and the systemd service is restarted

The application runs as a systemd service, managed by the operating system process supervisor, ensuring the application is automatically restarted in the event of a failure. Caddy is used as a reverse proxy in front of the Go binary, handling TLS certificate provisioning and renewal automatically via Let's Encrypt.

= DevOps Integration
To integrate DevOps pratices the project will be developed using a Version Control System (VCS) such as Git, which will allow shared ownership between team members developing the project, this approach also allows collaboration such that the team can iterate on several parts of the Backpack App seperately and deliver features at speed.

The project will also intergrate Automation which allows for performing repeated tasks relating to the delivery of the project such as deploying the app to the end once features have been approved and completed, and integrating risk reduction practices such as testing, dependency scanning etc.

In this project we will use Github for storing our code, which offers us a distributed version control system built on Git, out CI pipelines run on Github Actions, where our deployment action SSHs into our Hetzner VPS. The process running our binary on the server is managed using Sysemtd, a software suite for system and service management on Linux. Diagram showing the full integration can be found here, @integrated.

= Appendices
== Application Architecture Architecture <integrated>
#figure(image("docs/img/App Architecture.png", width: 80%), caption: "Backpack App Architecture with DevOps Pipeline")
