aws autoscaling update-auto-scaling-group --auto-scaling-group-name sgt-terraform-20180202210520490300000001 --desired-capacity 2 --profile playground --region us-east-1
echo "upsizing"
sleep 30
echo "30..."
sleep 30
echo "60..."
sleep 30
echo "90..."
sleep 30
echo "120..."
echo "downsizing"
aws autoscaling update-auto-scaling-group --auto-scaling-group-name sgt-terraform-20180202210520490300000001 --desired-capacity 1 --profile playground --region us-east-1
