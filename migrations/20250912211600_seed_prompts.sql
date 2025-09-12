-- +goose Up
-- +goose StatementBegin
INSERT INTO prompt_types (key, label, category) VALUES
-- About Me
('life_soundtrack', 'The soundtrack of my life is', 'About me'),
('friends_describe', 'Three words my friends use to describe me are', 'About me'),
('my_accent_origin', 'My accent or way of speaking comes from', 'About me'),
('podcast_title', 'If my life was a podcast, the first episode would be titled', 'About me'),
('noticed_when_talking', 'One thing people notice about me when I talk is', 'About me'),
('phrase_often_say', 'A phrase I probably say too often is', 'About me'),

-- Let’s Chat About
('best_story_heard', 'The best story I’ve ever heard is', 'Let''s chat about'),
('rabbit_hole', 'A random rabbit hole I could talk about for hours', 'Let''s chat about'),
('underrated_voice', 'The most underrated voice in music is', 'Let''s chat about'),
('debate_forever', 'Something I’ll always debate is', 'Let''s chat about'),
('cant_stay_quiet', 'If you bring this topic up, I can’t stay quiet', 'Let''s chat about'),
('last_laughed', 'The last thing that made me laugh out loud was', 'Let''s chat about'),

-- Getting Personal
('memory_shaped_me', 'A memory that shaped me is', 'Getting personal'),
('felt_alive', 'The moment I felt most alive was', 'Getting personal'),
('sound_takes_back', 'The sound that always takes me back is', 'Getting personal'),
('voice_never_forget', 'One voice I’ll never forget is', 'Getting personal'),
('learning_about_myself', 'Something I’m still learning about myself is', 'Getting personal'),
('life_turning_point', 'A turning point in my life was when', 'Getting personal'),

-- Playful & Fun
('impression_of', 'Do your best impression of', 'Playful & Fun'),
('sing_line', 'Sing one line from a song stuck in your head', 'Playful & Fun'),
('silly_belief', 'The silliest thing I believed as a kid', 'Playful & Fun'),
('tell_joke', 'Tell me a joke without laughing', 'Playful & Fun'),
('sound_effect_mood', 'Make a sound effect that describes your mood right now', 'Playful & Fun'),
('guess_noise', 'Guess this noise', 'Playful & Fun'),

-- Future & Dreams
('message_to_world', 'If I could broadcast one message to the world, I’d say', 'Future & Dreams'),
('dream_chasing', 'The dream I keep chasing is', 'Future & Dreams'),
('in_five_years', 'In five years, I hope I’m', 'Future & Dreams'),
('next_adventure', 'The next adventure on my list is', 'Future & Dreams'),
('ask_future_self', 'If I could speak to my future self, I’d ask', 'Future & Dreams'),
('never_stop_working', 'One thing I’ll never stop working towards is', 'Future & Dreams'),

-- Connection & Relationships
('make_people_heard', 'The way I make people feel heard is', 'Connection & Relationships'),
('small_act_of_love', 'A small act of love that means a lot to me is', 'Connection & Relationships'),
('quality_resonates', 'One quality that always resonates with me is', 'Connection & Relationships'),
('green_flag', 'A green flag for me is', 'Connection & Relationships'),
('if_i_like_you', 'If I like you, you’ll know because I', 'Connection & Relationships'),
('show_appreciation', 'The way I show appreciation is', 'Connection & Relationships'),

-- Haerd Exclusives (Voice-first)
('voice_changes', 'My voice sounds different when', 'Voice-first'),
('story_best_told', 'Here’s a story best told out loud', 'Voice-first'),
('describe_no_words', 'Describe yourself using only sounds (no words)', 'Voice-first'),
('my_laughter', 'What my laughter says about me', 'Voice-first'),
('vibe_sound_effect', 'If my vibe was a sound effect, it would be', 'Voice-first'),
('voice_every_day', 'One voice I’d love to hear every day is', 'Voice-first');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM prompt_types
WHERE key IN (
-- About Me
    'life_soundtrack','friends_describe','my_accent_origin','podcast_title','noticed_when_talking','phrase_often_say',

-- Let’s Chat About
    'best_story_heard','rabbit_hole','underrated_voice','debate_forever','cant_stay_quiet','last_laughed',

-- Getting Personal
    'memory_shaped_me','felt_alive','sound_takes_back','voice_never_forget','learning_about_myself','life_turning_point',

-- Playful & Fun
    'impression_of','sing_line','silly_belief','tell_joke','sound_effect_mood','guess_noise',

-- Future & Dreams
    'message_to_world','dream_chasing','in_five_years','next_adventure','ask_future_self','never_stop_working',

-- Connection & Relationships
    'make_people_heard','small_act_of_love','quality_resonates','green_flag','if_i_like_you','show_appreciation',

-- Haerd Exclusives (Voice-first)
    'voice_changes','story_best_told','describe_no_words','my_laughter','vibe_sound_effect','voice_every_day'
    );
-- +goose StatementEnd