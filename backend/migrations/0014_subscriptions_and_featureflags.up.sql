CREATE TABLE plans (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    features JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_subscriptions (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    plan_id TEXT NOT NULL REFERENCES plans(id),
    status TEXT NOT NULL CHECK (status IN ('active', 'expired')),
    valid_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO plans (id, name, features)
VALUES
    (
        'free',
        'Free',
        '{
          "basic_analysis": true,
          "history_days_limit": 30,
          "advanced_insights": false,
          "export_pdf": false,
          "export_json": false,
          "ci_comments": "limited",
          "baseline_sla": false,
          "team_management": false,
          "shared_projects": false,
          "advanced_trends": false
        }'::jsonb
    ),
    (
        'pro',
        'Pro',
        '{
          "basic_analysis": true,
          "history_days_limit": null,
          "advanced_insights": true,
          "export_pdf": true,
          "export_json": true,
          "ci_comments": true,
          "baseline_sla": true,
          "team_management": false,
          "shared_projects": false,
          "advanced_trends": false
        }'::jsonb
    ),
    (
        'team',
        'Team',
        '{
          "basic_analysis": true,
          "history_days_limit": null,
          "advanced_insights": true,
          "export_pdf": true,
          "export_json": true,
          "ci_comments": true,
          "baseline_sla": true,
          "team_management": true,
          "shared_projects": true,
          "advanced_trends": true
        }'::jsonb
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO user_subscriptions (user_id, plan_id, status)
SELECT id, 'free', 'active'
FROM users
ON CONFLICT (user_id) DO NOTHING;

CREATE OR REPLACE FUNCTION assign_default_subscription()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO user_subscriptions (user_id, plan_id, status)
    VALUES (NEW.id, 'free', 'active')
    ON CONFLICT (user_id) DO NOTHING;
    RETURN NEW;
END;
$$;

CREATE TRIGGER users_assign_default_subscription
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION assign_default_subscription();
