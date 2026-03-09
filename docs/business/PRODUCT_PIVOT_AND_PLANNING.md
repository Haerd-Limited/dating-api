# Haerd Product Pivot: Discovery & Planning

This pivot transforms Haerd from a psychological experiment into a high-intent utility tool. We are essentially building the **"Anti-Tinder"** — an app that prioritizes successful outcomes (dates) over endless dopamine loops (swiping).

---

## 1. The Vision

To eliminate "dating app fatigue" by enforcing intentionality. Haerd isn't a game to be played; it's a **logistics partner** designed to get users off the app and onto a high-quality date as efficiently as possible.

---

## 2. The Unique Value Propositions (The 5 Pillars)

- **The Equilibrium Protocol:** A strictly enforced 1:1 gender ratio (per country) ensures that no one is shouting into a void and no one is overwhelmed by low-quality noise.
- **The Date Concierge:** Integrated scheduling and venue selection to remove the "When and where?" friction that kills 50% of matches.
- **The Power of Two:** A hard cap of 2 active matches forces users to focus. You cannot "collect" people; you must either move forward or move on.
- **Deep Compatibility:** Data-driven matching using weighted question packs (importance-based) to ensure values align before the first "Hello."
- **Rich Expression:** Mandatory voice prompts and GIFs ensure the "vibe" is captured immediately. No more lazy, one-word text profiles.

---

## 3. Core User Personas

### 🎯 The Efficiency Seeker (25–40)

The professional who is tired of "Hey" and "How was your weekend?" They want to know if you're compatible by Tuesday and meet you by Thursday.

### 🎯 The Serious Dater (22–35)

Users who have deleted the "big" apps because they feel like a meat market. They value the 2-match limit because it guarantees the person they are talking to is actually paying attention.

---

## 4. The Product Workflow (The "Intentional Loop")

### Phase A: Onboarding

- **The Question Packs:** Users answer 20+ questions, marking each as "Irrelevant," "Important," or "Dealbreaker."
- **The Multimedia Profile:** Users must record at least 3 voice prompts and select a GIF (optional) that represents their personality. Photos are clear and visible.
- **The Gate:** If the gender ratio in their country is skewed, the user enters the Waiting Room.
- **Bypass Mechanic:** A male user can skip the queue immediately if he invites a female friend who completes a profile. Or men can pay to skip the queue (£10?).

### Phase B: Scarcity-Based Discovery

- **Relationship Compatibility %:** Displayed prominently on every profile.
- **The Active Match Cap:** Users can have 2 Active Matches in their inbox.
  - If a 3rd person matches, that match is "Locked."
  - The user sees a notification: *"You have a 96% match waiting! Unmatch an existing contact to start the conversation."* This forces an end to ghosting.

### Phase C: The Date Builder (Copied from Chaos)

- **Immediate Request:** At any point, a user can hit "Propose Date."
- **The Interface:** An in-app calendar shows mutual availability.
- **The Venue:** The app suggests curated local "Haerd-Approved" spots (bars, cafes, activities) based on the midpoint of their locations.
- **The Confirmation:** Once both accept, the date is "Booked" in-app.

---

## 5. Feature Requirements (MVP)

| Feature | Description | Progress |
|--------|-------------|----------|
| Relationship Compatibility | Question packs that allow us to calculate people's compatibility | ✅ |
| Voice/GIF Integration | Profile builder that rejects "empty" or "text-only" submissions | ✅ |
| The Queue System | Backend logic to pause new registrations based on gender ratio | |
| Lock/Unlock Logic | UI for the "Match Queue" when a user is at their 2-match limit | |
| Date Planner | API integration with Google/Apple Maps and Calendars | |

---

## 6. Strategic Monetization

**Suggestion from Andrew:** Hold off from monetisation for like a year or 2. Offer everything for free and build a massive user base then monetise. Which I think is a good idea.

### Premium Membership

- **Sort feed by Compatibility:** Pay and be able to see who you're most compatible with rather than swiping for ages.
- **Sort likes by Compatibility.**
- **See all likes** rather than one at a time.
- **More daily likes** — from 3 to 6 or 9?
- **3 Super likes a month:** Free users get 1 super like a month. (80+ compatibility can only be sent a super like)
- **The "Third and Fourth Slot":** 4 Active chats/matches instead of the usual 2.
- **Undo last skip (Rewind)**
- **See more profiles a day** — rather than usual 10? Profiles a day limit.

### One-off purchases

- **Superlikes** — Stand out from competition by sending superlike.
- **The "Queue Jump":** Pay a one-time fee to skip the Waiting Room (if the ratio isn't too extreme).

### Other revenue possibilities

- **Venue Partnerships:** (Long-term) Affiliate revenue from "Haerd-Approved" venues when users book through the app.

---

## 7. Risks & Mitigation

### The "Ghosting" Workaround

**Risk:** Users might stay in the "Active" section without replying.

**Possible Solution:** Implement a 48-hour "Inactivity Expiry" where a match automatically unmatches if no one speaks, freeing up the slot.

### The Ratio Stagnation

**Risk:** If a country is stuck at a 70/30 ratio, the waitlist might become too long.

**Solution:** Heavy referral incentives and localized "Join with a friend" marketing campaigns.

---

## 8. Non-MVP / Post-Launch Features

### Referral system

- Pay 10% of the referrer's referees who pay for premium.

### AI mentor/Host — "Hitch"

- Guides users throughout their chat like a blind dating show.
- It's optional. Host suggests to both users if they'd like to participate. If both accept, it begins.
- Paying users have access to an AI assistant that can help them with their Rizz.
- User has a personal mentor/Big bro that gives them tips/advice on how to come across as: Risky/Sexy, Romantic, funny.
- Give free users 3 uses. If they want more, they can upgrade.

### Rewarding good daters and punishing bad daters

- A predefined system or set of rules that define what good, mature dating looks like and what bad immature dating looks like.
  - **Bad:** Ghosting, taking days to reply.
  - **Good:** Quick replies, communicating properly after the reveal if you don't find them attractive.
- Every user on the app starts with a neutral dating score. Their actions and how they treat people affects it.
- People with very poor or very good dating scores are marked/highlighted for others to see.
- Others can also report/give a thumbs up after a date or reveal even if they don't end up together romantically. This will also affect their dating score.
- **Like Experian:** Paying users can see the reasons why their score went up or down and also see the scores of other people.
- Everyone will be able to see very good daters (green flag icon). But won't be able to see the bad daters.

### Voice calls

### Written Expression

- **Tagline / one-liner:** A witty or meaningful "about me."
- **Prompt-based text answers:** e.g., "Two truths and a lie," "Biggest risk I've taken," "Perfect weekend."
- **Quirky lists:** Top 3 movies, bucket list, pet peeves.
- **Username choice:** Serious, ironic, creative, or personal. Short caption space for each voice note — so the vibe is clear.

### Curation Expression — Favorites showcase

- Books
- Songs / playlists (Spotify/Apple integration)
- Quotes
- **What I'm into right now:** A space they can update regularly.
- **My vibe check:** 3 words they choose weekly.
- Interest/Hobbies tags — #anime #bouldering

### Interactive Expression

- **Mini-games in profiles:** e.g., "Swipe to guess: which of these is their lie?"
- **Polls / icebreakers:** "What should I try next weekend?"
- **Q&A mode:** Allow others to ask anonymous fun questions (with voice replies).
- **Challenge prompts:** Weekly fun challenge (tell a joke, sing 10s of your fav song, etc.) to reveal personality.

### Personality Signaling Features

- **Values / beliefs selectors:** Faith, intentions, lifestyle choices (but with fun UX).
- **Humor test:** Show them a meme and let them voice-react. Their reaction itself says a lot. e.g. maybe allow users to add a youtube reel/video link which would be embedded into their profile.
- **Conversation starters section:** Users pre-select the kinds of convos they like (deep, silly, philosophical, playful).
- **MBTI / Enneagram / custom archetypes (optional)** — but in a playful way, not heavy.

### Special periods/events

### Personality Signaling Features (continued)

- **Audio bio intro** (like a podcast opening).
- **Choose an avatar style:** cartoon silhouettes, abstract icons, AI-generated vibe art.
- **Mood color of the week:** they choose a color to represent their current vibe.
- **Daily thought snippet** (1-sentence status that updates frequently).

**👉 The big principle:** make the "scrolling experience" on Haerd feel like entering someone's mind/room — their humor, music, voice, quirks — not just looking at their face.

### Social / Relational Expression

- **Endorsements from friends:** Friends can leave one-liners about you ("Best storyteller I know").
- **Voice shoutouts:** Allow friends to record a short compliment/testimonial for your profile — via sending a link and without downloading the app.
- **Shared vibes:** Show overlaps in music taste, books, or habits with potential matches.

### Implement Monetisation

- Group voice rooms (like speed dating nights)
- Consistency challenge (Win)
- AI convo starters
- Scheduling feature and notifications
- Usage Stats/Metrics on profiles
- More ways to show your interests
- **Haerd Web**
- Opening up pre-registration & initial launch to Americans (+1)
- **Date builder**
