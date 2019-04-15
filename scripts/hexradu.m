hexradius=2
result=[];
for n=1:1000

theta=rand()*360;
maxx=(0.866+(abs(mod(abs(theta),60)-30)/30)*0.134)*hexradius;
radius=rand()*maxx;
x=radius*cos(theta); 
y=radius*sin(theta);
result(n,:)=[x,y];

end

figure;
scatter(result(:,1),result(:,2))