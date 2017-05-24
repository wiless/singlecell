% Script evaluates the distance between the associated BS & UE, and SINR
%
info=stable(:,[1,7,10:11])
bsids=info(:,2)';
count=1
for k=bsids
ind=find(bslocations(:,1)==k);
info(count,5:6)=bslocations(ind,2:3);
usloc=info(count,2:3);
bsloc=info(count,5:6);
d=(usloc-bsloc);
info(count,7)=norm(d);
info(count,8)=stable(count,8);
count=count+1;
end