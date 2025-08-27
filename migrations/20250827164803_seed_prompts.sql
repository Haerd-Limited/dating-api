-- +goose Up
-- +goose StatementBegin
INSERT INTO prompt_types (key, label, category) VALUES
-- About Me
('greatest_strength', 'My greatest strength', 'About me'),
('unusual_skills', 'Unusual skills', 'About me'),
('goal_this_year', 'This year, I really want to', 'About me'),
('dating_me_is_like', 'Dating me is like', 'About me'),
('random_fact_love', 'A random fact I love is', 'About me'),
('simple_pleasures', 'My simple pleasures', 'About me'),
('typical_sunday', 'Typical Sunday', 'About me'),
('way_to_win_me', 'The way to win me over is', 'About me'),
('irrational_fear', 'My most irrational fear', 'About me'),
('life_goal', 'A life goal of mine', 'About me'),
('recent_discovery', 'I recently discovered that', 'About me'),

-- Let’s Chat About
('change_my_mind', 'Change my mind about', 'Let''s chat about'),
('teach_me', 'Teach me something about', 'Let''s chat about'),
('debate_topic', 'Let''s debate this topic', 'Let''s chat about'),
('i_bet_you_cant', 'I bet you can''t', 'Let''s chat about'),
('guess_this', 'Try to guess this about me', 'Let''s chat about'),
('travel_tips', 'Give me travel tips for', 'Let''s chat about'),
('pick_topic', 'I''ll pick the topic if you start the conversation', 'Let''s chat about'),
('same_page', 'Let''s make sure we''re on the same page about', 'Let''s chat about'),
('agree_disagree', 'Do you agree or disagree that', 'Let''s chat about'),
('thing_to_know', 'The one thing I''d love to know about you is', 'Let''s chat about'),
('leave_comment', 'You should leave a comment if', 'Let''s chat about'),

-- Getting Personal
('dont_hate_me', 'Don''t hate me if I', 'Getting personal'),
('what_if', 'What if I told you that', 'Getting personal'),
('dont_date_me_if', 'You should *not* go out with me if', 'Getting personal'),
('key_to_my_heart', 'The key to my heart is', 'Getting personal'),
('dorkiest_thing', 'The dorkiest thing about me is', 'Getting personal'),
('one_thing', 'The one thing you should know about me is', 'Getting personal'),
('wont_shut_up', 'I won''t shut up about', 'Getting personal'),
('love_language', 'My Love Language is', 'Getting personal'),
('geek_out_on', 'I geek out on', 'Getting personal'),
('if_loving_wrong', 'If loving this is wrong, I don''t want to be right', 'Getting personal'),

-- Your World
('friend_group_role', 'In my friend group, I’m the one who', 'Your World'),
('feel_like_myself', 'Where I go when I want to feel a little more like myself', 'Your World'),
('you_never_know', 'You’d never know it, but I', 'Your World'),
('talking_all_night', 'I could stay up all night talking about', 'Your World'),
('pet_thinks', 'Something my pet thinks about me', 'Your World'),
('family_award', 'An award my family would give me', 'Your World'),
('in_element', 'I’m in my element when', 'Your World'),
('holiday_must', 'It’s not a holiday unless', 'Your World'),
('before_meet_listen', 'Before we meet, you should listen to', 'Your World'),
('kindest_thing', 'The kindest thing someone has ever done for me', 'Your World'),

-- My Type
('looking_for', 'I''m looking for', 'My type'),
('want_someone', 'I want someone who', 'My type'),
('id_fall_for', 'I''d fall for you if', 'My type'),
('non_negotiable', 'Something that''s non-negotiable for me is', 'My type'),
('brag_about_you', 'I''ll brag about you to my friends if', 'My type'),
('green_flags', 'Green flags I look out for', 'My type'),
('good_relationship', 'The hallmark of a good relationship is', 'My type'),
('weirdly_attracted', 'I''m weirdly attracted to', 'My type'),
('get_along_if', 'We''ll get along if', 'My type'),
('same_type_of_weird', 'We''re the same type of weird if', 'My type'),
('all_i_ask', 'All I ask is that you', 'My type'),

-- Self-Care
('therapy_taught_me', 'Therapy recently taught me', 'Self-care'),
('relaxation_is', 'To me, relaxation is', 'Self-care'),
('need_advice', 'When I need advice, I go to', 'Self-care'),
('therapist_would_say', 'My therapist would say I', 'Self-care'),
('i_unwind', 'I unwind by', 'Self-care'),
('friends_ask_me', 'My friends ask me for advice about', 'Self-care'),
('self_care_routine', 'My self-care routine is', 'Self-care'),
('beat_my_blues', 'I beat my blues by', 'Self-care'),
('last_journal', 'My last journal entry was about', 'Self-care'),
('cry_in_car_song', 'My cry-in-the-car song is', 'Self-care'),
('last_cried_happy', 'The last time I cried happy tears was', 'Self-care'),
('feel_supported', 'I feel most supported when', 'Self-care'),
('hype_myself_up', 'I hype myself up by', 'Self-care'),
('boundary', 'A boundary of mine is', 'Self-care'),

-- Date Vibes
('together_we_could', 'Together, we could', 'Date vibes'),
('best_way_to_ask', 'The best way to ask me out is by', 'Date vibes'),
('first_round', 'First round is on me if', 'Date vibes'),
('best_spot', 'I know the best spot in town for', 'Date vibes'),
('order_for_table', 'What I order for the table', 'Date vibes'),

-- Story Time
('worst_idea', 'Worst idea I''ve ever had', 'Story time'),
('biggest_risk', 'Biggest risk I''ve taken', 'Story time'),
('weirdest_gift', 'Weirdest gift I''ve given or received', 'Story time'),
('travel_story', 'Best travel story', 'Story time'),
('never_again', 'One thing I''ll never do again', 'Story time'),
('never_have_i', 'Never have I ever', 'Story time'),
('most_spontaneous', 'Most spontaneous thing I''ve done', 'Story time'),
('biggest_date_fail', 'My biggest date fail', 'Story time'),
('two_truths_lie', 'Two truths and a lie', 'Story time'),

-- Voice First
('bff_reasons', 'My BFF''s reasons for why you should date me', 'Voice-first'),
('fav_film_line', 'My favourite line from a film', 'Voice-first'),
('musical_talent', 'Proof I have musical talent', 'Voice-first'),
('shower_thought', 'A thought I recently had in the shower', 'Voice-first'),
('soundtrack', 'Apparently, my life''s soundtrack is', 'Voice-first'),
('celebrity_impression', 'My best celebrity impression', 'Voice-first'),
('pronounce_name', 'How to pronounce my name', 'Voice-first'),
('say_hi', 'Saying “Hi!” in all the languages I know', 'Voice-first'),
('wish_more_knew', 'I wish more people knew', 'Voice-first'),
('guess_song', 'Guess the song', 'Voice-first'),
('set_up_punchline', 'I''ll give you the set-up; you guess the punchline', 'Voice-first'),
('dad_joke', 'My best dad joke', 'Voice-first'),
('quick_rant', 'A quick rant about', 'Voice-first');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM prompt_types
WHERE key IN (
  -- About me
  'greatest_strength','unusual_skills','goal_this_year','dating_me_is_like','random_fact_love',
  'simple_pleasures','typical_sunday','way_to_win_me','irrational_fear','life_goal','recent_discovery',

  -- Let's chat about
  'change_my_mind','teach_me','debate_topic','i_bet_you_cant','guess_this','travel_tips',
  'pick_topic','same_page','agree_disagree','thing_to_know','leave_comment',

  -- Getting personal
  'dont_hate_me','what_if','dont_date_me_if','key_to_my_heart','dorkiest_thing','one_thing',
  'wont_shut_up','love_language','geek_out_on','if_loving_wrong',

  -- Your World
  'friend_group_role','feel_like_myself','you_never_know','talking_all_night','pet_thinks',
  'family_award','in_element','holiday_must','before_meet_listen','kindest_thing',

  -- My type
  'looking_for','want_someone','id_fall_for','non_negotiable','brag_about_you','green_flags',
  'good_relationship','weirdly_attracted','get_along_if','same_type_of_weird','all_i_ask',

  -- Self-care
  'therapy_taught_me','relaxation_is','need_advice','therapist_would_say','i_unwind',
  'friends_ask_me','self_care_routine','beat_my_blues','last_journal','cry_in_car_song',
  'last_cried_happy','feel_supported','hype_myself_up','boundary',

  -- Date vibes
  'together_we_could','best_way_to_ask','first_round','best_spot','order_for_table',

  -- Story time
  'worst_idea','biggest_risk','weirdest_gift','travel_story','never_again','never_have_i',
  'most_spontaneous','biggest_date_fail','two_truths_lie',

  -- Voice-first
  'bff_reasons','fav_film_line','musical_talent','shower_thought','soundtrack',
  'celebrity_impression','pronounce_name','say_hi','wish_more_knew','guess_song',
  'set_up_punchline','dad_joke','quick_rant'
);
-- +goose StatementEnd
